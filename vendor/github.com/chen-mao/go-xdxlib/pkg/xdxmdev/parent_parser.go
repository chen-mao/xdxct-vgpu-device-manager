package xdxmdev

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/chen-mao/go-xdxlib/pkg/xdxpci"
)

type ParentDevice struct {
	*xdxpci.XDXCTPCIDevice
	mdevPaths map[string]string
}

func NewParentDevice(devicePath string) (*ParentDevice, error) {
	reg := regexp.MustCompile(`Type Name: (\w+)`)
	xdxDevice, err := newXDXCTPCIDeviceFromPath(devicePath)
	if err != nil {
		return nil, fmt.Errorf("failed to construct XDXCT PCI device: %v", err)
	}
	if xdxDevice == nil {
		// Not a XDXCT device
		return nil, err
	}

	paths, err := filepath.Glob(fmt.Sprintf("%s/mdev_supported_types/xgv-XGV_V0_*/name", xdxDevice.Path))
	if err != nil {
		return nil, fmt.Errorf("unable to get files in mdev_supported_types directory: %v", err)
	}
	mdevTypesMap := make(map[string]string)
	for _, path := range paths {
		name, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("unable to read file %s: %v", path, err)
		}
		mdevTypeStr := strings.TrimSpace(string(name))
		matches := reg.FindStringSubmatch(mdevTypeStr)
		if len(matches) > 1 {
			extracted := matches[1]
			mdevTypesMap[extracted] = filepath.Dir(path)
		} else {
			return nil, fmt.Errorf("unable to parse mdev_type name for mdev %s", mdevTypeStr)
		}
	}
	return &ParentDevice{
		xdxDevice,
		mdevTypesMap,
	}, nil
}

func newXDXCTPCIDeviceFromPath(devicePath string) (*xdxpci.XDXCTPCIDevice, error) {
	root := filepath.Dir(devicePath)
	address := filepath.Base(devicePath)
	return xdxpci.New(xdxpci.WithPCIDevicesRoot(root)).
		GetGPUByPciBusID(address)
}

// CreateMDEVDevice creates a mediated device (vGPU) on the parent GPU
func (pd *ParentDevice) CreateMDEVDevice(mdevType string, uuid string) error {
	mdevPath, ok := pd.mdevPaths[mdevType]
	if !ok {
		return fmt.Errorf("unable to create mdev %s: mde not supported by parent device %s", mdevType, pd.Address)
	}

	file, err := os.OpenFile(filepath.Join(mdevPath, "create"), os.O_WRONLY|os.O_SYNC, 0200)
	if err != nil {
		return fmt.Errorf("unable to open create file: %v", err)
	}

	_, err = file.WriteString(uuid)
	if err != nil {
		return fmt.Errorf("unable to create mdev: %v", err)
	}

	return nil
}

// IsMDEVTypeSupported checks if the mdevType is supported by the GPU
func (pd *ParentDevice) IsMDEVTypeSupported(mdevType string) bool {
	_, found := pd.mdevPaths[mdevType]
	return found
}

func (pd *ParentDevice) GetAvailableMDEVInstances(mdevType string) (int, error) {
	mdevPath, ok := pd.mdevPaths[mdevType]
	if !ok {
		return -1, nil
	}
	available, err := os.ReadFile(filepath.Join(mdevPath, "available_instances"))
	if err != nil {
		return -1, fmt.Errorf("unable to read available instances: %v", err)
	}

	availableInstances, err := strconv.Atoi(strings.TrimSpace(string(available)))
	if err != nil {
		return -1, fmt.Errorf("unable to convert available instances to an int: %v", err)
	}
	return availableInstances, nil
}
