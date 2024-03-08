package app

import (
	"bufio"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/chen-mao/go-xdxlib/pkg/xdxpci"
	"github.com/chen-mao/xdxct-vgpu-device-manager/pkg/types"
	"github.com/chen-mao/xdxct-vgpu-device-manager/pkg/vgpu"
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

func WalkSelectedVGPUConfigForEachGPU(vGPUConfig VGPUConfigSpecSlice, f func(VGPUConfigSpec, int) error) error {
	xdxpci := xdxpci.New()
	gpus, err := xdxpci.GetGPUs()
	log.Debugf("gpu on node: %d", len(gpus))
	if err != nil {
		return fmt.Errorf("error enumerating GPUs: %v", err)
	}
	for _, vc := range vGPUConfig {
		if vc.DeviceFilter == nil {
			log.Debugf("Walking VGPUConfig for (devices=%v)", vc.Devices)
		} else {
			log.Debugf("Walking VGPUConfig for (device-filter=%v, devices=%v)", vc.DeviceFilter, vc.Devices)
		}

		for i, gpu := range gpus {
			// to do
			deviceID := types.NewDeviceID(gpu.Device, gpu.Vendor)
			// if !vc.MatchDeviceFilter(deviceID) {
			// 	continue
			// }

			if !vc.MatchDevices(i) {
				continue
			}

			log.Debugf("GPU %v: %v", i, deviceID)

			err = f(vc, i)
			if err != nil {
				return nil
			}
		}
	}

	return nil
}

// AssertVGPUConfig asserts that the selected vGPU config is applied to the node
func AssertVGPUConfig(vGPUConfig VGPUConfigSpecSlice) error {
	xdxpci := xdxpci.New()
	gpus, err := xdxpci.GetGPUs()
	if err != nil {
		return fmt.Errorf("error get gpus info: %v", err)
	}
	matched := make([]bool, len(gpus))
	err = WalkSelectedVGPUConfigForEachGPU(vGPUConfig, func(vs VGPUConfigSpec, index int) error {
		configManager := vgpu.NewXdxlibVGPUConfigManager()
		currentVGPUConfig, err := configManager.GetVGPUConfig(index)
		if err != nil {
			return fmt.Errorf("error get vGPU config: %v", err)
		}

		log.Debugf("Asserting vGPU config: %v", vs.VGPUDevices)
		if currentVGPUConfig.Equals(vs.VGPUDevices) {
			log.Debugf("Skipping -- already set to desired value")
			matched[index] = true
			return nil
		}

		matched[index] = false
		return nil
	})

	if err != nil {
		return err
	}

	for _, match := range matched {
		if !match {
			return fmt.Errorf("not all GPUs match the specified config")
		}
	}

	return nil
}

// ApplyVGPUConfig applies the selected vGPU config to the node
func ApplyVGPUConfig(VGPUConfig VGPUConfigSpecSlice) error {
	return WalkSelectedVGPUConfigForEachGPU(VGPUConfig, func(vs VGPUConfigSpec, index int) error {
		configManager := vgpu.NewXdxlibVGPUConfigManager()
		currentVGPUConfig, err := configManager.GetVGPUConfig(index)
		if err != nil {
			return fmt.Errorf("error getting vGPU config: %v", err)
		}
		log.Debugf("Current vGPU config: %v", currentVGPUConfig)

		if currentVGPUConfig.Equals(vs.VGPUDevices) {
			log.Debugf("Skipping -- already set to desired value")
			return nil
		}

		log.Debugf("Updating vGPU config: %v", vs.VGPUDevices)
		err = configManager.SetVGPUConfig(index, vs.VGPUDevices)
		if err != nil {
			return fmt.Errorf("error setting VGPU config: %v", err)
		}
		return nil
	})
}
