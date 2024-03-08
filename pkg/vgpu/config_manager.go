package vgpu

import (
	"fmt"

	"github.com/chen-mao/go-xdxlib/pkg/xdxmdev"
	"github.com/chen-mao/xdxct-vgpu-device-manager/internal/xdxlib"
	"github.com/chen-mao/xdxct-vgpu-device-manager/pkg/types"
	"github.com/google/uuid"
)

type Manager interface {
	GetVGPUConfig(gpu int) (types.VGPUConfig, error)
	SetVGPUConfig(gpu int, config types.VGPUConfig) error
	ClearVGPUConfig(gpu int) error
}

type xdxlibVGPUConfigManager struct {
	xdxlib xdxlib.Interface
}

func NewXdxlibVGPUConfigManager() Manager {
	return &xdxlibVGPUConfigManager{
		xdxlib.New(),
	}
}

// GetVGPUConfig gets the 'VGPUConfig' currently applied to a GPU at a particular index
func (cm *xdxlibVGPUConfigManager) GetVGPUConfig(gpu int) (types.VGPUConfig, error) {
	parentGPUDevice, err := cm.xdxlib.Xdxpci.GetGPUByIndex(gpu)
	if err != nil {
		return nil, fmt.Errorf("error getting device at index '%d': '%v'", gpu, err)
	}

	vGPUDevices, err := cm.xdxlib.Xdxmdev.GetAllMediatedDevices()
	if err != nil {
		return nil, fmt.Errorf("error getting all vgpu devices: '%v'", err)
	}
	vGpuConfigs := types.VGPUConfig{}
	for _, vGPUDevice := range vGPUDevices {
		if parentGPUDevice.Address == vGPUDevice.Parent.Address {
			vGpuConfigs[vGPUDevice.MDEVType]++
		}
	}
	return vGpuConfigs, nil
}

// SetVGPUConfig applies the selected `VGPUConfig` to a GPU at a particular index if it is not already applied
func (cm *xdxlibVGPUConfigManager) SetVGPUConfig(gpu int, config types.VGPUConfig) error {
	parentGPUDevice, err := cm.xdxlib.Xdxpci.GetGPUByIndex(gpu)
	if err != nil {
		return fmt.Errorf("error getting device ay index '%d','%v'", gpu, err)
	}
	allDevicesInfo, err := cm.xdxlib.Xdxmdev.GetAllParentDevices()
	if err != nil {
		return fmt.Errorf("error getting all parent devices: %v", err)
	}

	var currentDevices []*xdxmdev.ParentDevice
	for _, p := range allDevicesInfo {
		if p.Device == parentGPUDevice.Device {
			currentDevices = append(currentDevices, p)
		}
	}

	if len(currentDevices) == 0 {
		return fmt.Errorf(" no parent devices found for GPU at index: %d", gpu)
	}
	for key := range config {
		if !currentDevices[0].IsMDEVTypeSupported(key) {
			return fmt.Errorf("vGPU type %s is not support on GPU (indev=%d, address=%s)", key, gpu, parentGPUDevice.Address)
		}
	}

	err = cm.ClearVGPUConfig(gpu)
	if err != nil {
		return fmt.Errorf("error clearing VGPUConfig: %v", err)
	}
	for key, value := range config {
		remainingToCreate := value
		for _, currentDevice := range currentDevices {
			if remainingToCreate == 0 {
				break
			}
			supported := currentDevice.IsMDEVTypeSupported(key)
			if !supported {
				return fmt.Errorf("error get available vGPU instances: %v", err)
			}

			available, err := currentDevice.GetAvailableMDEVInstances(key)
			if err != nil {
				return fmt.Errorf("unable to get available gpu instances: %v", err)
			}
			if available <= 0 {
				continue
			}
			numToCreate := min(remainingToCreate, available)
			for i := 0; i < remainingToCreate; i++ {
				err = currentDevice.CreateMDEVDevice(key, uuid.New().String())
				if err != nil {
					return fmt.Errorf("unable to create %s vGPU device on parent device %s: %v", key, currentDevice.Address, err)
				}
			}
			remainingToCreate -= numToCreate
		}

		if remainingToCreate > 0 {
			return fmt.Errorf("failed to create mdev vgpu deivce %s", key)
		}
	}
	return nil
}

func (cm *xdxlibVGPUConfigManager) ClearVGPUConfig(gpu int) error {
	pciDeviceInfo, err := cm.xdxlib.Xdxpci.GetGPUByIndex(gpu)
	if err != nil {
		return fmt.Errorf("error getting device ay index '%d','%v'", gpu, err)
	}
	vGPUDevInfos, err := cm.xdxlib.Xdxmdev.GetAllMediatedDevices()
	if err != nil {
		return fmt.Errorf("error getting all vGPU devices: %v", err)
	}

	for _, vgpuDevInfo := range vGPUDevInfos {
		if pciDeviceInfo.Address == vgpuDevInfo.Parent.Address {
			err := vgpuDevInfo.Delete()
			if err != nil {
				return fmt.Errorf("error deleting %s vgpu with id %s: %v", vgpuDevInfo.MDEVType, vgpuDevInfo.UUID, err)
			}
		}
	}
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
