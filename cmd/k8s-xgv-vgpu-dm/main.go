package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const (
	resourceNodes                    = "nodes"
	vGPUConfigLabel                  = "xdxct.com/vgpu-config"
	cliName                          = "xgv-vgpu-dm"
	kubevirt_device_plugin_namespace = "kube-system"
	kubevirt_device_plugin_Label     = "name=xdxct-kubevirt-dp-ds"
)

var (
	kubeconfigFlag        string
	nodeNameFlag          string
	namespaceFlag         string
	configFileFlag        string
	defaultVGPUConfigFlag string
)

type SyncableVGPUConfig struct {
	cond           *sync.Cond
	mutex          sync.Mutex
	current        string
	lastVGPUConfig string
}

func NewSyncableVGPUConfig() *SyncableVGPUConfig {
	var m SyncableVGPUConfig
	m.cond = sync.NewCond(&m.mutex)
	return &m
}

func (m *SyncableVGPUConfig) Set(value string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.current = value
	if m.current != "" {
		m.cond.Broadcast()
	}
}

func (m *SyncableVGPUConfig) Get() string {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.lastVGPUConfig == m.current {
		m.cond.Wait()
	}
	m.lastVGPUConfig = m.current
	return m.lastVGPUConfig
}

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

func notifyVGPUConfigChangesFromNode(clientset *kubernetes.Clientset, vGPUConfig *SyncableVGPUConfig) chan struct{} {
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
				vGPUConfig.Set(obj.(*corev1.Node).Labels[vGPUConfigLabel])
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				oldLabel := oldObj.(*corev1.Node).Labels[vGPUConfigLabel]
				newLabel := newObj.(*corev1.Node).Labels[vGPUConfigLabel]
				if oldLabel != newLabel {
					vGPUConfig.Set(newLabel)
				}
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

	// to do: delete line 177~182
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

	vGPUConfig := NewSyncableVGPUConfig()

	stopch := notifyVGPUConfigChangesFromNode(clientset, vGPUConfig)
	defer close(stopch)

	//Apply initial vGPU configuration
	selectedConfig, err := getNodeLabel(clientset)
	if err != nil {
		return fmt.Errorf("unable to get vGPU config label: %v", err)
	}
	if selectedConfig == "" {
		log.Infof("No vGPU config specified for node. Proceeding with default config: %s", defaultVGPUConfigFlag)
		selectedConfig = defaultVGPUConfigFlag
	} else {
		selectedConfig = vGPUConfig.Get()
	}

	log.Infof("Updating to vGPU config: %s", selectedConfig)
	err = updateConfig(clientset, selectedConfig)
	if err != nil {
		log.Errorf("ERROR: %v", err)
	} else {
		log.Infof("Successfully updated to vGPU config: %s", selectedConfig)
	}

	for {
		log.Infof("Waiting for change to %s label", vGPUConfigLabel)
		value := vGPUConfig.Get()
		log.Infof("Updating t vGPU config: %s", value)
		err = updateConfig(clientset, value)
		if err != nil {
			log.Errorf("ERROR: %v", err)
			continue
		}
		log.Infof("Successfully update to vGPU config: %s", value)
	}
}

func getNodeLabel(clientset *kubernetes.Clientset) (string, error) {
	node, err := clientset.CoreV1().Nodes().Get(context.TODO(), nodeNameFlag, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("unable to get node obj: %v", err)
	}
	value, ok := node.Labels[vGPUConfigLabel]
	if !ok {
		return "", nil
	}
	return value, nil
}

func updateConfig(clientset *kubernetes.Clientset, selectedConfig string) error {
	// to do add validator components
	// nvidia采用的删除label, operator 会shutdown validation和kubevirt-device-plugin的组件，接着再去重新调用组件。
	// 这里是删除了pod, daemonset 会重启pod达到重启的效果。
	log.Info("Restart all GPU Component in Kubernetes by disabling their nodeSelector labels")
	err := deleteGPUComponent(clientset)
	if err != nil {
		fmt.Printf("Unable to delete GPU component: %v", err)
		return err
	}

	log.Info("Applying the selected vGPU device configuration to the node")
	err = applyConfig(selectedConfig)
	if err != nil {
		fmt.Printf("Unable to apply config %s: %v", selectedConfig, err)
		return err
	}
	return nil
}

// to do: delete pod by set label
// to do: add validation
func deleteGPUComponent(clientset *kubernetes.Clientset) error {
	Pods, err := clientset.CoreV1().Pods(kubevirt_device_plugin_namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: kubevirt_device_plugin_Label,
		FieldSelector: fmt.Sprintf("spec.nodeName=%s", nodeNameFlag),
	})
	if err != nil {
		return fmt.Errorf("unable to get pod object: %v", err)
	}
	for _, pod := range Pods.Items {
		clientset.CoreV1().Pods(kubevirt_device_plugin_namespace).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("unable to get pod object: %v", err)
		}
	}

	log.Infof("waiting for kubevirt-device-plugin to delete")
	err = waitForPodDeletion(clientset)
	if err != nil {
		return fmt.Errorf("failed to delete po:%v", err)
	}

	return nil
}

func waitForPodDeletion(clientset *kubernetes.Clientset) error {
	timeOut := time.After(120 * time.Second)
	timeInterval := time.Second * 2
	for {
		podList, err := clientset.CoreV1().Pods(kubevirt_device_plugin_namespace).List(context.TODO(), metav1.ListOptions{
			LabelSelector: kubevirt_device_plugin_Label,
			FieldSelector: fmt.Sprintf("spec.nodeName=%s", nodeNameFlag),
		})
		if apierrors.IsNotFound(err) {
			return nil
		}
		if len(podList.Items) == 0 {
			return nil
		}

		select {
		case <-time.After(timeInterval):
		case <-timeOut:
			return fmt.Errorf("failed to delete pod: %v", err)
		}
	}
}

func applyConfig(config string) error {
	args := []string{
		"-v",
		"apply",
		"-f", configFileFlag,
		"-c", config,
	}
	cmd := exec.Command(cliName, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
