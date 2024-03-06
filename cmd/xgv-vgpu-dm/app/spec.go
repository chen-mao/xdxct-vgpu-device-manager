package app

import (
	"fmt"
	"reflect"

	"github.com/chen-mao/xdxct-vgpu-device-manager/pkg/types"
)

type Spec struct {
	Version     string                         `json:"version" yaml:"version"`
	VGPUConfigs map[string]VGPUConfigSpecSlice `json:"vgpu-configs,omitempty" yaml:"vgpu-configs,omitempty"`
}

type VGPUConfigSpec struct {
	DeviceFilter interface{}      `json:"version" yaml:"version"`
	Devices      interface{}      `json:"devices" yaml:"devices,flow"`
	VGPUDevices  types.VGPUConfig `json:"vgpu-devices" yaml:"vgpu-devices"`
}

type VGPUConfigSpecSlice []VGPUConfigSpec

// to do
func (vc *VGPUConfigSpec) MatchDeviceFilter(deviceID types.DeviceID) bool {
	return true
}

func (vc *VGPUConfigSpec) MatchAllDevices() bool {
	switch devices := vc.Devices.(type) {
	case string:
		return devices == "all"
	}
	return false
}

func (vc *VGPUConfigSpec) MatchDevices(index int) bool {
	fmt.Printf("type: %v\n", reflect.TypeOf(vc.Devices))
	switch devices := vc.Devices.(type) {
	case []interface{}:
		for _, d := range devices {
			if index == d {
				return true
			}
		}
	}
	return vc.MatchAllDevices()
}
