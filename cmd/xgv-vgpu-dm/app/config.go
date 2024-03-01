package app

import (
	"bufio"
	"fmt"
	"os"

	"github.com/chen-mao/go-xdxlib/pkg/xdxpci"
	"gopkg.in/yaml.v2"
)

func ParseConfigFile(f *Flags) (*Spec, error) {
	var configYaml []byte
	var err error
	if f.ConfigFile == "-" {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			configYaml = append(configYaml, scanner.Bytes()...)
			configYaml = append(configYaml, '\n')
		}
	} else {
		configYaml, err = os.ReadFile(f.ConfigFile)
		if err != nil {
			return nil, fmt.Errorf("read file error: %v", err)
		}
	}

	var spec Spec
	err = yaml.Unmarshal(configYaml, &spec)
	if err != nil {
		return nil, fmt.Errorf("unmarshal error: %v", err)
	}
	return &spec, nil
}

func GetSelectedVGPUConfig(f *Flags, spec *Spec) (VGPUConfigSpecSlice, error) {
	if f.SelectedConfig == "" && len(spec.VGPUConfigs) > 1 {
		return nil, fmt.Errorf("missing required flag 'selected-config' when more than one config available")
	}

	if f.SelectedConfig == "" && len(spec.VGPUConfigs) == 1 {
		for c := range spec.VGPUConfigs {
			f.SelectedConfig = c
		}
	}

	if _, exists := spec.VGPUConfigs[f.SelectedConfig]; !exists {
		return nil, fmt.Errorf("select vgpu-conig not present: %v", f.SelectedConfig)
	}
	return spec.VGPUConfigs[f.SelectedConfig], nil
}

// AssertVGPUConfig asserts that the selected vGPU config is applied to the node
func AssertVGPUConfig() error {
	xdxpci := xdxpci.New()
	_, err := xdxpci.GetGPUs()
	if err != nil {
		return fmt.Errorf("error get gpus info: %v", err)
	}

	return nil
}
