package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	cli "github.com/urfave/cli/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const (
	resourceNodes = "nodes"
)

var (
	kubeconfigFlag        string
	nodeNameFlag          string
	namespaceFlag         string
	configFileFlag        string
	defaultVGPUConfigFlag string
)

func main() {
	app := cli.NewApp()
	app.Before = validationFlags
	app.Action = start

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "kubeconfig",
			Value:       "",
			Usage:       "the absolute path to kubeconfig file",
			Destination: &kubeconfigFlag,
			EnvVars:     []string{"KUBECONFIG"},
		},
		&cli.StringFlag{
			Name:        "nodeName",
			Aliases:     []string{"n"},
			Value:       "",
			Usage:       "watch the name of label",
			Destination: &nodeNameFlag,
			EnvVars:     []string{"NODENAME"},
		},
		&cli.StringFlag{
			Name:        "namespace",
			Aliases:     []string{"ns"},
			Value:       "",
			Usage:       "the namespace where the GPU components deployed",
			Destination: &namespaceFlag,
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

func validationFlags(c *cli.Context) error {
	if nodeNameFlag == "" {
		return fmt.Errorf("invalid <node-name> flag: must not be empty string")
	}
	if namespaceFlag == "" {
		return fmt.Errorf("invalid <namespace> flag: must not be empty string")
	}
	if configFileFlag == "" {
		return fmt.Errorf("invalid <configFileFlag> flag: must not be empty string")
	}
	if defaultVGPUConfigFlag == "" {
		return fmt.Errorf("invalid <default-VGPU-Config> flag: must not be empty string")
	}
	return nil
}

func notifyVGPUConfigChangesFromNode(clientset *kubernetes.Clientset) chan struct{} {
	lw := cache.NewListWatchFromClient(
		clientset.CoreV1().RESTClient(),
		resourceNodes,
		corev1.NamespaceAll,
		fields.OneTermEqualSelector("metadata.name", nodeNameFlag),
	)

	_, controller := cache.NewInformer(
		lw, &corev1.Node{},
		10*time.Second,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				fmt.Println(obj.(*corev1.Node).Labels["mchen"])
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				fmt.Println(oldObj.(*corev1.Node).Labels["mchen"])
				fmt.Println(newObj.(*corev1.Node).Labels["mchen"])
			},
		},
	)

	stopch := make(chan struct{})

	go controller.Run(stopch)

	return stopch
}

func start(c *cli.Context) error {
	var config *rest.Config
	var err error

	// to do: delete line 128 ~ 133
	if home := homedir.HomeDir(); home != "" {
		kubeconfigFlag = *flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfigFlag = *flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	if config, err = rest.InClusterConfig(); err != nil {
		if config, err = clientcmd.BuildConfigFromFlags("", kubeconfigFlag); err != nil {
			return fmt.Errorf("error building kubernetes clientcmd config: %s", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("error building kubernetes clientset from config: %s", err)
	}

	stopch := notifyVGPUConfigChangesFromNode(clientset)
	defer close(stopch)

	select {}

	return nil
}
