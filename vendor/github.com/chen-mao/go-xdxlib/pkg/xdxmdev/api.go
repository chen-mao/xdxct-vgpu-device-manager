package xdxmdev

// Interface allows us to get a list of XDXCT MDEV (vGPU) and parent devices
// GetAllMediatedDevices return 'all vGPUConfig' currently applied to the 'index' gpu
type Interface interface {
	GetAllMediatedDevices() ([]*Device, error)
	GetAllParentDevices() ([]*ParentDevice, error)
}
