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

func (cm *xdxlibVGPUConfigManager) GetVGPUConfig(gpu int) (types.VGPUConfig, error) {
	return nil, nil
}

func (cm *xdxlibVGPUConfigManager) SetVGPUConfig(gpu int, config types.VGPUConfig) error {
	fmt.Println("--- start set vgpu config ---")
	pciDeviceInfo, err := cm.xdxlib.Xdxpci.GetGPUByIndex(gpu)
	if err != nil {
		return fmt.Errorf("error getting device ay index '%d','%v'", gpu, err)
	}
	allDevicesInfo, err := cm.xdxlib.Xdxmdev.GetAllParentDevices()
	if err != nil {
		return fmt.Errorf("error getting all parent devices: %v", err)
	}

	var currentDevices []*xdxmdev.ParentDevice
	for _, p := range allDevicesInfo {
		if p.Device == pciDeviceInfo.Device {
			currentDevices = append(currentDevices, p)
		}
	}

	if len(currentDevices) == 0 {
		return fmt.Errorf(" no parent devices found for GPU at index: %d", gpu)
	}
	for key := range config {
		if !currentDevices[0].IsMDEVTypeSupported(key) {
			return fmt.Errorf("vGPU type %s is not support on GPU (indev=%d, address=%s)", key, gpu, pciDeviceInfo.Address)
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
			// to do check available_mdev_instances

			for i := 0; i < remainingToCreate; i++ {
				err = currentDevice.CreateMDEVDevice(key, uuid.New().String())
				if err != nil {
					return fmt.Errorf("unable to create %s vGPU device on parent device %s: %v", key, currentDevice.Address, err)
				}
			}
			remainingToCreate--
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
