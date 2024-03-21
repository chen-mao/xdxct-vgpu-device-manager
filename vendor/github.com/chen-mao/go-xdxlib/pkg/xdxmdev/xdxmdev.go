package xdxmdev

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
)

const (
	mdevParentRoot = "/sys/class/mdev_bus"
	mdevDeviceRoot = "/sys/bus/mdev/devices"
)

type xdxmdev struct {
	mdevParentRoot string
	mdevDeviceRoot string
}

func New() Interface {
	return &xdxmdev{
		mdevParentRoot: mdevParentRoot,
		mdevDeviceRoot: mdevDeviceRoot,
	}
}

// GetAllDevices returns all XDXCT mdev (vGPU) devices on the system
func (m *xdxmdev) GetAllMediatedDevices() ([]*Device, error) {
	deviceDirs, err := os.ReadDir(m.mdevDeviceRoot)
	if err != nil {
		return nil, fmt.Errorf("unable to read PCI bus devices: %v", err)
	}
	var xdxdevices []*Device
	for _, deviceDir := range deviceDirs {
		xdxdevice, err := NewMediatedDevice(m.mdevDeviceRoot, deviceDir.Name())
		if err != nil {
			return nil, fmt.Errorf("error constructing xdxct MDEV device: %v", err)
		}
		if xdxdevice == nil {
			continue
		}
		xdxdevices = append(xdxdevices, xdxdevice)
	}
	return xdxdevices, nil
}

// GetAllParentDevices returns all XDXCT Parent PCI devices on the system
func (m *xdxmdev) GetAllParentDevices() ([]*ParentDevice, error) {
	deviceDirs, err := os.ReadDir(m.mdevParentRoot)
	if err != nil {
		return nil, fmt.Errorf("unable to read PCI bus devices: %v", err)
	}

	var xdxdevices []*ParentDevice
	for _, deviceDir := range deviceDirs {
		devicePath := path.Join(m.mdevParentRoot, deviceDir.Name())
		xdxdevice, err := NewParentDevice(devicePath)
		if err != nil {
			return nil, fmt.Errorf("error constructing xdxct parent device: %v", err)
		}
		if xdxdevice == nil {
			continue
		}
		xdxdevices = append(xdxdevices, xdxdevice)
	}

	addressToID := func(address string) uint64 {
		address = strings.ReplaceAll(address, ":", "")
		address = strings.ReplaceAll(address, ".", "")
		id, _ := strconv.ParseUint(address, 16, 64)
		return id
	}

	sort.Slice(xdxdevices, func(i, j int) bool {
		return addressToID(xdxdevices[i].Address) < addressToID(xdxdevices[j].Address)
	})

	return xdxdevices, nil
}
