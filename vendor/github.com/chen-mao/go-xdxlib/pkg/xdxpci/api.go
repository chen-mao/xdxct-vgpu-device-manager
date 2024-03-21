package xdxpci

// Interface allows us to get a list of all XDXCT PCI devices
type Interface interface {
	GetGPUs() ([]*XDXCTPCIDevice, error)
	GetGPUByIndex(int) (*XDXCTPCIDevice, error)
	GetGPUByPciBusID(string) (*XDXCTPCIDevice, error)
}
