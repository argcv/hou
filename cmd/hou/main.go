package main

import (
	"os"

	"github.com/argcv/stork/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/argcv/hou"
)

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
		Short: "Host Objects Ultra-lightly",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			verbose, err := cmd.Flags().GetBool("verbose")

			if verbose {
				log.SetLevel(log.DEBUG)
				log.Debugf("verbose mode: ON")
			}

			port, err := cmd.Flags().GetInt("port")
			if err != nil {
				return err
			}

			defaultFile, err := cmd.Flags().GetString("default")
			if err != nil {
				return
			}

			indexFile, err := cmd.Flags().GetString("index")
			if len(indexFile) == 0 {
				indexFile = defaultFile
			}

			baseDir, err := cmd.Flags().GetString("base")
			if err != nil {
				return err
			}

			debug, err := cmd.Flags().GetBool("debug")
			if err != nil {
				return err
			}

			proxy, err := cmd.Flags().GetString("proxy")
			if err != nil {
				return err
			}

			h := hou.New()

			h.Basedir = baseDir
			h.DefaultFile = defaultFile
			h.IndexFile = indexFile
			h.Port = port
			h.Debug = debug
			h.Proxy = proxy

			log.Infof("Starting:\n%v", h.ConfigTable())
			return h.Run()
		},
	}

	args.PersistentFlags().String("default", "index.html", "default file")
	args.PersistentFlags().String("index", "", "index file")
	args.PersistentFlags().String("base", ".", "base dir")
	args.PersistentFlags().IntP("port", "p", 6789, "port")
	args.PersistentFlags().String("proxy", "", "remote proxy")

	args.PersistentFlags().BoolP("debug", "d", false, "debug mode")
	args.PersistentFlags().BoolP("verbose", "v", false, "verbose log")

	if err := args.Execute(); err != nil {
		log.Debugf("Execute Failed: %v", err.Error())
	}
}
