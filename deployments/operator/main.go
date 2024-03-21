package main

import (
	"log"
	"os"

	cli "github.com/urfave/cli/v2"
)

var (
	kubeconfig            string
	nodeNameFlag          string
	configFileFlag        string
	defaultVGPUConfigFlag string
)

func main() {
	app := cli.NewApp()
	app.Action = start

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "kubeconfig",
			Value:       "",
			Usage:       "the absolute path to kubeconfig file",
			Destination: &kubeconfig,
			EnvVars:     []string{"KUBECONFIG"},
		},
		&cli.StringFlag{
			Name:        "nodeName",
			Aliases:     []string{"n"},
			Value:       "",
			Usage:       "watch the name of label",
			Destination: &nodeNameFlag,
			EnvVars:     []string{"NODEName"},
		},
		&cli.StringFlag{
			Name:        "namespace",
			Aliases:     []string{"ns"},
			Value:       "",
			Usage:       "the namespace where the GPU components deployed",
			Destination: &nodeNameFlag,
			EnvVars:     []string{"NAMESPACE"},
		},
		&cli.StringFlag{
			Name:        "configFile",
			Aliases:     []string{"f"},
			Value:       "",
			Usage:       "the absolute path to vGPU config file",
			Destination: &configFileFlag,
			EnvVars:     []string{"CONFIGFILE"},
		},
		&cli.StringFlag{
			Name:        "default-vgpu-config",
			Aliases:     []string{"d"},
			Value:       "",
			Usage:       "the default vGPU config to use if no label is set",
			Destination: &defaultVGPUConfigFlag,
			EnvVars:     []string{"DEFAULTVGPUCONFIG"},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.SetOutput(os.Stderr)
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
}

func start(c *cli.Context) error {
	return nil
}
