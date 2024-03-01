package app

import "github.com/chen-mao/xdxct-vgpu-device-manager/pkg/types"

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
