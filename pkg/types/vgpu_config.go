package types

type VGPUConfig map[string]int

// // Contains checks if the provided 'vgpuType' is part of the 'VGPUConfig'.
// func (v VGPUConfig) Contains(vgpuType string) bool {
// 	if _, exists := v[vgpuType]; !exists {
// 		return false
// 	}
// 	return v[vgpuType] > 0
// }

// // Equals checks if two 'VGPUConfig's are equal.
// // Equality is determined by comparing the vGPU types contained in each 'VGPUConfig'.
// func (v VGPUConfig) Equals(config VGPUConfig) bool {
// 	if len(v) != len(config) {
// 		fmt.Printf("len v is: %d\n", len(v))
// 		fmt.Printf("len config is: %d\n", len(config))
// 		return false
// 	}
// 	for k, v := range v {
// 		if !config.Contains(k) {
// 			return false
// 		}
// 		if v != config[k] {
// 			return false
// 		}
// 	}
// 	return true
// }
