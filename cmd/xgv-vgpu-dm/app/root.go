package app

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var Verbose bool

var rootCmd = &cobra.Command{
	Use:     os.Args[0],
	Version: "0.1.0",
	Short:   "xgv vgpu device manager tool",
	Long:    "This is the xgv vgpu device manager tool.",
}

func InitConfig() {
	if Verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel((log.InfoLevel))
	}
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: false,
	})
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", os.Getenv("VERBOSE") == "true", "Enable verbose logging")
	cobra.OnInitialize(InitConfig)
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
