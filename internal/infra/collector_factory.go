package infra

import (
	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

type CollectorFactory struct{}

func (f *CollectorFactory) CreateCollectors() []any {
	var collectors []any

	collectors = append(collectors, NewCPUMemoryCollector())

	if hasNvidiaGPU() {
		collectors = append(collectors, NewNvidiaGPUCollector())
	}
	// TODO: Implement AMD GPU detection and collection in the future
	// else if hasAMDGPU() {
	//     collectors = append(collectors, NewAMDGPUCollector())
	// }

	return collectors
}

func hasNvidiaGPU() bool {
	ret := nvml.Init()
	if ret != nvml.SUCCESS {
		return false
	}
	defer nvml.Shutdown()

	count, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		return false
	}

	return count > 0
}

// TODO: Implement AMD GPU detection in the future
// func hasAMDGPU() bool {
//     // Implementation to detect AMD GPU
//     return false
// }
