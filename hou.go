package main

import (
	"fmt"
	"github.com/argcv/webeh/log"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path"
)

func IsValidFile(index string, files ...string) string {

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
				sf := IsValidFile(index, path.Join(f, index))
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

func main() {
	viper.SetConfigName("hou")
	viper.SetEnvPrefix("hou")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.hou/")
	viper.AddConfigPath("/etc/")
	if conf := os.Getenv("HOU_CFG"); conf != "" {
		viper.SetConfigFile(conf)
	}

	args := &cobra.Command{
		Use:   "hou",
		Short: "hou hou hou",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			verbose, err := cmd.Flags().GetBool("verbose")

			if verbose {
				log.SetLevel(log.DEBUG)
				log.Debugf("verbose mode: ON")
			}

			port, err := cmd.Flags().GetInt("port")
			defaultFile, err := cmd.Flags().GetString("default")
			indexFile, err := cmd.Flags().GetString("index")
			if len(indexFile) == 0 {
				indexFile = defaultFile
			}

			debug, err := cmd.Flags().GetBool("debug")

			if debug {
				// debug mode
				gin.SetMode(gin.DebugMode)
				log.Infof("Mode: Debug")
			} else {
				gin.SetMode(gin.ReleaseMode)
				log.Infof("Mode: Release")
			}

			r := gin.Default()

			r.GET("/*file", func(c *gin.Context) {
				file := path.Clean(path.Join(".", path.Clean(c.Param("file"))))
				log.Debugf("Requested Path: %s", file)
				fileIn := IsValidFile(indexFile, file, defaultFile)

				if len(fileIn) > 0 {
					//data, err := ioutil.ReadFile(fileIn)
					if err != nil {
						log.Warnf("Read failed: %v", err.Error())
						c.String(501, err.Error())
					} else {
						c.File(fileIn)
					}
				} else {
					c.String(404, "File Not Found")
				}

			})

			listenStr := fmt.Sprintf(":%v", port)
			log.Infof("Listening: %v", listenStr)
			log.Infof("Default File: %v", defaultFile)
			r.Run(listenStr) // listen and serve on 0.0.0.0:8080

			return
		},
	}

	args.PersistentFlags().String("default", "index.html", "default file")
	args.PersistentFlags().String("index", "", "index file")
	args.PersistentFlags().IntP("port", "p", 6789, "port")

	args.PersistentFlags().BoolP("debug", "d", false, "debug mode")
	args.PersistentFlags().BoolP("verbose", "v", false, "verbose log")

	if err := args.Execute(); err != nil {
		log.Debugf("Execute Failed: %v", err.Error())
	}
}
