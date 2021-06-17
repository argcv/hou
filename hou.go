package hou

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/argcv/stork/log"
	"github.com/gin-gonic/gin"
	"github.com/olekukonko/tablewriter"
)

type Hou struct {
	Basedir     string
	Port        int
	DefaultFile string
	IndexFile   string

	Debug bool

	BodyNotFound string
}

func New() *Hou {
	return &Hou{
		Basedir:      ".",
		DefaultFile:  "index.html",
		BodyNotFound: "File Not Found",
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
	}
	table := tablewriter.NewWriter(buff)
	table.SetHeader([]string{"option", "value"})
	for _, v := range data {
		table.Append(v)
	}
	table.Render() // Send output
	return buff.String()
}

func (h *Hou) Run() error {
	if h.Debug {
		// debug mode
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	indexFile := h.GetIndexFile()

	r.GET("/*file", func(c *gin.Context) {
		file := path.Clean(path.Join(h.Basedir, path.Clean(c.Param("file"))))
		log.Debugf("Requested Path: %s", file)
		fileIn := ScanValidFile(indexFile, file, h.DefaultFile)

		if len(fileIn) > 0 {
			//c.File(fileIn)
			http.ServeFile(c.Writer, c.Request, fileIn)
			//http.ServeContent(c.Writer, c.Request, "name", time.Now(), os.Open(fileIn))
		} else {
			c.String(404, h.BodyNotFound)
		}
	})

	return r.Run(fmt.Sprintf(":%v", h.Port))
}

func ScanValidFile(index string, files ...string) string {

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
				sf := ScanValidFile(index, path.Join(f, index))
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
