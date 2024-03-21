package xdxpci

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	// PCIDevicesRoot represents base path for all pci devices under sysfs
	PCIDevicesRoot = "/sys/bus/pci/devices"
	// PCIXDXCTVendorID represents PCI vendor id for XDXCT
	PCIXDXCTVendorID uint16 = 0x1eed
	// PCIVgaControllerClass represents the PCI class for VGA Controllers
	PCIVgaControllerClass uint32 = 0x030000
)

// XDXCTPCIDevice represents a PCI device for an XDXCT product
// TO ADD XDX PCI Device Info
// ClassName  string
// DeviceName string
// resource  MemoryResources
type XDXCTPCIDevice struct {
	Path       string
	Address    string
	Vendor     uint16
	Class      uint32
	Device     uint16
	Driver     string
	IommuGroup int
	NumaNode   int
	Config     *ConfigSpace
}

func (xp *XDXCTPCIDevice) IsGPU() bool {
	return xp.Class == PCIVgaControllerClass
}

type xdxpci struct {
	logger         logger
	pciDevicesRoot string
	pcidbPath      string
}

func New(opts ...Option) Interface {
	n := &xdxpci{}
	for _, opt := range opts {
		opt(n)
	}
	if n.logger == nil {
		n.logger = &simpleLogger{}
	}
	if n.pciDevicesRoot == "" {
		n.pciDevicesRoot = PCIDevicesRoot
	}
	return n
}

// Option defines a function for passing options to the New() call
type Option func(*xdxpci)

func WithLogger(logger logger) Option {
	return func(n *xdxpci) {
		n.logger = logger
	}
}

func WithPCIDevicesRoot(path string) Option {
	return func(n *xdxpci) {
		n.pcidbPath = path
	}
}

func (p xdxpci) GetGPUByPciBusID(address string) (*XDXCTPCIDevice, error) {
	parentDevicePath := filepath.Join(p.pciDevicesRoot, address)

	vendor, err := os.ReadFile(path.Join(parentDevicePath, "vendor"))
	if err != nil {
		return nil, fmt.Errorf("unable to read pci device vendor id for %s: %v", address, err)
	}
	vendorStr := strings.TrimSpace(string(vendor))
	vendorID, err := strconv.ParseUint(vendorStr, 0, 16)
	if err != nil {
		return nil, fmt.Errorf("unable to convert vendor string to uint16: %v", vendorStr)
	}
	if uint16(vendorID) != PCIXDXCTVendorID {
		return nil, nil
	}

	class, err := os.ReadFile(path.Join(parentDevicePath, "class"))
	if err != nil {
		return nil, fmt.Errorf("unable to read pci device class for %s: %v", address, err)
	}
	classStr := strings.TrimSpace(string(class))
	classId, err := strconv.ParseUint(classStr, 0, 32)
	if err != nil {
		return nil, fmt.Errorf("unable to convert class string to uint16: %v", classStr)
	}

	device, err := os.ReadFile(path.Join(parentDevicePath, "device"))
	if err != nil {
		return nil, fmt.Errorf("unable to read PCI device id for %s: %v", address, err)
	}
	deviceStr := strings.TrimSpace(string(device))
	deviceID, err := strconv.ParseUint(deviceStr, 0, 16)
	if err != nil {
		return nil, fmt.Errorf("unable to convert device string to uint16: %v", deviceStr)
	}

	driver, err := filepath.EvalSymlinks(path.Join(parentDevicePath, "driver"))
	if err == nil {
		driver = filepath.Base(driver)
	} else if os.IsNotExist(err) {
		driver = ""
	} else {
		return nil, fmt.Errorf("unable to detect driver for %s: %v", address, err)
	}

	var iommuGroup int64
	iommu, err := filepath.EvalSymlinks(path.Join(parentDevicePath, "iommu_group"))
	if err == nil {
		iommuGroupStr := strings.TrimSpace(filepath.Base(iommu))
		iommuGroup, err = strconv.ParseInt(iommuGroupStr, 0, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to convert iommu_group string to int64: %v", iommuGroupStr)
		}
	} else if os.IsNotExist(err) {
		iommuGroup = -1
	} else {
		return nil, fmt.Errorf("unable to detect iommu_group for %s: %v", address, err)
	}

	numa, err := os.ReadFile(path.Join(parentDevicePath, "numa_node"))
	if err != nil {
		return nil, fmt.Errorf("unable to read PCI NUMA node for %s: %v", address, err)
	}
	numaStr := strings.TrimSpace(string(numa))
	numaNode, err := strconv.ParseInt(numaStr, 0, 64)
	if err != nil {
		return nil, fmt.Errorf("unable to convert NUMA node string to int64: %v", numaNode)
	}

	config := &ConfigSpace{
		Path: path.Join(parentDevicePath, "config"),
	}

	return &XDXCTPCIDevice{
		Path:       parentDevicePath,
		Address:    address,
		Vendor:     uint16(vendorID),
		Class:      uint32(classId),
		Device:     uint16(deviceID),
		Driver:     driver,
		IommuGroup: int(iommuGroup),
		NumaNode:   -1,
		Config:     config,
	}, nil
}

func (p *xdxpci) GetAllDevices() ([]*XDXCTPCIDevice, error) {
	deviceDirs, err := os.ReadDir(p.pciDevicesRoot)
	if err != nil {
		return nil, fmt.Errorf("unable to read PCI bus devices: %v", err)
	}
	var xdxdevices []*XDXCTPCIDevice
	for _, deviceDir := range deviceDirs {
		deviceAddress := deviceDir.Name()
		xdxdevice, err := p.GetGPUByPciBusID(deviceAddress)
		if err != nil {
			return nil, fmt.Errorf("error constructing xdxct pci device %s: %v", deviceAddress, err)
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

// GetGPUs returns all XDXCT GPU devices on the system
func (p *xdxpci) GetGPUs() ([]*XDXCTPCIDevice, error) {
	devices, err := p.GetAllDevices()
	if err != nil {
		return nil, fmt.Errorf("error getting all xdxct devices: %v", err)
	}
	var filtered []*XDXCTPCIDevice
	for _, d := range devices {
		if d.IsGPU() {
			filtered = append(filtered, d)
		}
	}
	return filtered, nil
}

func (p xdxpci) GetGPUByIndex(i int) (*XDXCTPCIDevice, error) {
	gpus, err := p.GetGPUs()
	if err != nil {
		return nil, fmt.Errorf("error getting all gpus: %v", err)
	}
	if i < 0 || i > len(gpus) {
		return nil, fmt.Errorf("invalid index '%d'", i)
	}
	return gpus[i], nil
}
