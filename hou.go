package hou

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/argcv/stork/log"
	"github.com/gin-gonic/gin"
	"github.com/olekukonko/tablewriter"
)

type Hou struct {
	Basedir     string
	Port        int
	DefaultFile string
	IndexFile   string
	Proxy       string

	Debug bool

	BodyNotFound string
}

func New() *Hou {
	return &Hou{
		Basedir:      ".",
		DefaultFile:  "index.html",
		BodyNotFound: "File Not Found",
		Proxy:        "",
	}
}

func (h *Hou) GetIndexFile() string {
	if len(h.IndexFile) == 0 {
		return h.DefaultFile
	} else {
		return h.IndexFile
	}
}

func (h *Hou) String() string {
	lb := len(h.BodyNotFound)
	if lb > 5 {
		lb = 5
	}
	return fmt.Sprintf("Hou[:%v][Index(%v), Default(%v) Debug(%v) Not Found(%v...)]",
		h.Port, h.GetIndexFile(), h.DefaultFile, h.Debug, h.BodyNotFound[:lb])
}

func (h Hou) ConfigTable() string {
	buff := bytes.NewBuffer(nil)

	lb := len(h.BodyNotFound)
	if lb > 20 {
		lb = 20
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "127.0.0.1"
	}

	data := [][]string{
		{"Endpoint", fmt.Sprintf("http://%s:%d", hostname, h.Port)},
		{"Debug", fmt.Sprint(h.Debug)},
		{"Index", h.GetIndexFile()},
		{"Default", h.DefaultFile},
		{"Not Found", h.BodyNotFound[:lb]},
		{"Proxy", h.Proxy},
	}
	table := tablewriter.NewWriter(buff)
	table.SetHeader([]string{"option", "value"})
	for _, v := range data {
		table.Append(v)
	}
	table.Render() // Send output
	return buff.String()
}

func (h *Hou) handlerLocal(r gin.IRouter) {
	indexFile := h.GetIndexFile()

	r.GET("/*file", func(c *gin.Context) {
		file := path.Clean(path.Join(h.Basedir, path.Clean(c.Param("file"))))
		log.Debugf("Requested Path: %s", file)
		fileIn := ScanLocalValidFile(indexFile, file, h.DefaultFile)

		if len(fileIn) > 0 {
			//c.File(fileIn)
			http.ServeFile(c.Writer, c.Request, fileIn)
			//http.ServeContent(c.Writer, c.Request, "name", time.Now(), os.Open(fileIn))
		} else {
			c.String(404, h.BodyNotFound)
		}
	})
}

func (h *Hou) handlerRemote(r gin.IRouter) {
	indexFile := path.Clean("/" + h.GetIndexFile())

	proxy := h.Proxy
	if !strings.HasPrefix(proxy, "http://") && !strings.HasPrefix(proxy, "https://") {
		proxy = "http://" + proxy
	}
	if strings.HasSuffix(proxy, "/") {
		proxy = proxy[:len(proxy)-1]
	}

	log.Infof("using proxy %v", proxy)
	client := &http.Client{
		Timeout: 300 * time.Second,
	}
	r.GET("/*file", func(c *gin.Context) {
		file := c.Param("file")
		//file := path.Clean(path.Join(h.Basedir, path.Clean(c.Param("file"))))
		url := proxy + file
		log.Infof("[%v] requesting... %v", file, url)
		resp, err := client.Get(url)
		if err == nil && resp.StatusCode != 404 {
			c.Status(resp.StatusCode)
			defer resp.Body.Close()
			_, err = io.Copy(c.Writer, resp.Body)
			if err != nil {
				log.Errorf("request %v failed: %v", url, err)
			}
			return
		}

		url = proxy + indexFile

		resp, err = client.Get(url)
		if err != nil {
			c.Status(502)
			log.Errorf("request %v failed: %v", url, err)
			return
		}
		c.Status(resp.StatusCode)
		defer resp.Body.Close()
		_, err = io.Copy(c.Writer, resp.Body)
		if err != nil {
			log.Errorf("request %v failed: %v", url, err)
		}
	})
}

func (h *Hou) Run() error {
	if h.Debug {
		// debug mode
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	if len(h.Proxy) > 0 {
		h.handlerRemote(r)
	} else {
		h.handlerLocal(r)
	}

	return r.Run(fmt.Sprintf(":%v", h.Port))
}

func ScanLocalValidFile(index string, files ...string) string {

	log.Debugf("Scanning...")
	for _, f := range files {
		log.Debugf("[%v] will be checked...", f)
	}
	for _, f := range files {
		log.Debugf("[%v] checking...", f)
		if stat, err := os.Stat(f); err == nil {
			log.Debugf("[%v] exists...", f)
			if stat.IsDir() {
				log.Debugf("[%v] is dir...", f)
				sf := ScanLocalValidFile(index, path.Join(f, index))
				if len(sf) > 0 {
					return sf
				}
			} else {
				// is file
				return f
			}
		}
	}
	return ""
}
