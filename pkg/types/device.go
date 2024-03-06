package types

type DeviceID uint32

func NewDeviceID(device, vendor uint16) DeviceID {
	return DeviceID(uint32(device)<<16 | uint32(vendor))
}
