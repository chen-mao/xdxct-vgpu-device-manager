package types

type VGPUConfig map[string]int

// Contains checks if the provided 'vgpuType' is part of the 'VGPUConfig'.
func (vc *VGPUConfig) Contains(vgpuType string) bool {
	if _, exists := (*vc)[vgpuType]; !exists {
		return false
	}
	return (*vc)[vgpuType] > 0
}

// Equals checks if two 'VGPUConfig's are equal.
// Equality is determined by comparing the vGPU types contained in each 'VGPUConfig'.
func (vc *VGPUConfig) Equals(config VGPUConfig) bool {
	if len(*vc) != len(config) {
		return false
	}
	for k, v := range *vc {
		if !config.Contains(k) {
			return false
		}
		if v != config[k] {
			return false
		}
	}
	return true
}
