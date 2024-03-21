package xdxmdev

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type Device struct {
	Path       string
	UUID       string
	MDEVType   string
	Driver     string
	IommuGroup int
	Parent     ParentDevice
}

// Delete deletes a mediated device (vGPU)
func (d *Device) Delete() error {
	removefile, err := os.OpenFile(filepath.Join(d.Path, "remove"), os.O_WRONLY|os.O_SYNC, 0200)
	if err != nil {
		return fmt.Errorf("unable to open remove file: %v", err)
	}
	_, err = removefile.WriteString("1")
	if err != nil {
		return fmt.Errorf("unable to delete mdev: %v", err)
	}
	return nil
}

// NewDevice constructs a Device, which represents an xdxct mdev (vGPU) device
func NewMediatedDevice(root string, uuid string) (*Device, error) {
	path := path.Join(root, uuid)

	m, err := newMdev(path)
	if err != nil {
		return nil, err
	}

	parent, err := NewParentDevice(m.parentDevicePath())
	if err != nil {
		return nil, fmt.Errorf("error getting mdev type: %v", err)
	}
	if parent == nil {
		return nil, nil
	}

	mdevType, err := m.Type()
	if err != nil {
		return nil, fmt.Errorf("error get mdev type: %v", err)
	}

	driver, err := m.Driver()
	if err != nil {
		return nil, fmt.Errorf("error detecting driver: %v", err)
	}

	iommu_group, err := m.IommuGroup()
	if err != nil {
		return nil, fmt.Errorf("error detecting Iommu Group: %v", err)
	}

	device := Device{
		Path:       path,
		UUID:       uuid,
		MDEVType:   mdevType,
		Driver:     driver,
		IommuGroup: int(iommu_group),
		Parent:     *parent,
	}
	return &device, nil
}

type mdev string

func newMdev(devicePath string) (mdev, error) {
	mdevDir, err := filepath.EvalSymlinks(devicePath)
	if err != nil {
		return "", fmt.Errorf("error resolving symlink for %s: %v", devicePath, err)
	}
	return mdev(mdevDir), nil
}

// parentDevicePath() return "/sys/devices/pci0000:00/0000:00:01.1/<pcu-id>"
func (m *mdev) parentDevicePath() string {
	return path.Dir(string(*m))
}

func (m *mdev) resolve(target string) (string, error) {
	resolved, err := filepath.EvalSymlinks(path.Join(string(*m), target))
	if err != nil {
		return "", fmt.Errorf("error resolving %q: %v", target, err)
	}
	return resolved, nil
}

func (m *mdev) Type() (string, error) {
	reg := regexp.MustCompile(`Type Name: (\w+)`)

	mdevTypeDir, err := m.resolve(string("mdev_type"))
	if err != nil {
		return "", err
	}
	mdevTypeName, err := os.ReadFile(path.Join(mdevTypeDir, "name"))
	if err != nil {
		return "", fmt.Errorf("unable to read mdev_type name for mdev %s: %v", mdevTypeName, err)
	}

	mdevTypeStr := strings.TrimSpace(string(mdevTypeName))

	matches := reg.FindStringSubmatch(mdevTypeStr)
	if len(matches) > 1 {
		extracted := matches[1]
		return extracted, nil
	} else {
		return "", fmt.Errorf("unable to parse mdev_type name for mdev %s", mdevTypeName)
	}
}

func (m *mdev) Driver() (string, error) {
	driver, err := m.resolve(string("driver"))
	if err != nil {
		return "", err
	}

	return filepath.Base(driver), nil
}

func (m *mdev) IommuGroup() (int64, error) {
	IommuGroup, err := m.resolve(string("iommu_group"))
	if err != nil {
		return -1, err
	}

	IommuGroupStr := strings.TrimSpace(filepath.Base(IommuGroup))
	IommuGroupInt, err := strconv.ParseInt(IommuGroupStr, 0, 64)
	if err != nil {
		return -1, fmt.Errorf("unable to convert iommu_group string to int64: %v", err)
	}

	return IommuGroupInt, nil
}
