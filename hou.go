package hou

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/argcv/stork/log"
	"github.com/gin-gonic/gin"
	"github.com/olekukonko/tablewriter"
)

type Hou struct {
	Basedir      string
	Port         int
	DefaultFile  string
	IndexFile    string
	Proxy        string
	ProxyHeaders map[string]string

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
	if len(h.ProxyHeaders) > 0 {
		for k, v := range h.ProxyHeaders {
			data = append(data, []string{"Proxy Headers", strings.Join([]string{k, v}, ":")})
		}
	}
	table := tablewriter.NewWriter(buff)
	table.SetHeader([]string{"option", "value"})
	table.SetAutoMergeCells(true)
	table.AppendBulk(data)
	table.Render() // Send output
	return buff.String()
}

func (h *Hou) handlerLocal(r gin.IRouter) {
	indexFile := h.GetIndexFile()

	r.Any("/*file", func(c *gin.Context) {
		file := path.Clean(path.Join(h.Basedir, path.Clean(c.Param("file"))))
		log.Debugf("Requested Path: %s", file)
		fileIn := ScanLocalValidFile(indexFile, file, h.DefaultFile)

		if len(fileIn) > 0 {
			// c.File(fileIn)
			http.ServeFile(c.Writer, c.Request, fileIn)
			// http.ServeContent(c.Writer, c.Request, "name", time.Now(), os.Open(fileIn))
		} else {
			c.String(404, h.BodyNotFound)
		}
	})
}

func (h *Hou) handlerRemote(r gin.IRouter) {
	proxy := h.Proxy
	if !strings.HasPrefix(proxy, "http://") && !strings.HasPrefix(proxy, "https://") {
		proxy = "http://" + proxy
	}
	if strings.HasSuffix(proxy, "/") {
		proxy = proxy[:len(proxy)-1]
	}

	host := ""
	myURL, err := url.Parse(proxy)
	if err == nil {
		host = myURL.Host
	}

	buildRequest := func(c *gin.Context, file string) *http.Request {
		targetURL := proxy + file

		// log.Infof("[%v] requesting... %v", file, targetURL)

		req := c.Request.Clone(c.Request.Context())
		req.URL, _ = url.Parse(targetURL)
		req.RequestURI = ""

		for k, v := range h.ProxyHeaders {
			req.Header.Set(k, v)
		}
		if host != "" {
			req.Host = host
		}
		return req
	}

	copyHeader := func(c *gin.Context, header http.Header) {
		for k, v := range header {
			c.Writer.Header().Del(k)
			for _, vx := range v {
				c.Writer.Header().Add(k, vx)
			}
		}
	}

	log.Infof("using proxy %v", proxy)
	client := httpClient
	r.Any("/*file", func(c *gin.Context) {
		file := c.Param("file")
		resp, err := client.Do(buildRequest(c, file))
		if err == nil {
			c.Status(resp.StatusCode)
			defer resp.Body.Close()

			copyHeader(c, resp.Header)

			_, err = io.Copy(c.Writer, resp.Body)
			if err != nil {
				log.Errorf("request %v failed: %v", file, err)
			}
			return
		}

		log.Warnf("request %v failed: %v", file, err)
		c.Status(502)
		return
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
