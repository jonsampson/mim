package infra

import (
	"fmt"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/jonsampson/mim/internal/domain"
)

type NvidiaGPUCollector struct {
	*BaseCollector[domain.GPUMetrics]
}

func NewNvidiaGPUCollector() *NvidiaGPUCollector {
	collector := &NvidiaGPUCollector{}
	collector.BaseCollector = NewBaseCollector(collector.getMetrics)
	return collector
}

func (c *NvidiaGPUCollector) getMetrics() (domain.GPUMetrics, error) {
	ret := nvml.Init()
	if ret != nvml.SUCCESS {
		return domain.GPUMetrics{}, fmt.Errorf("failed to initialize NVML: %v", ret)
	}
	defer nvml.Shutdown()

	count, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		return domain.GPUMetrics{}, fmt.Errorf("failed to get device count: %v", ret)
	}

	if count == 0 {
		return domain.GPUMetrics{}, fmt.Errorf("no NVIDIA GPUs found")
	}

	device, ret := nvml.DeviceGetHandleByIndex(0)
	if ret != nvml.SUCCESS {
		return domain.GPUMetrics{}, fmt.Errorf("failed to get device handle: %v", ret)
	}

	utilization, ret := device.GetUtilizationRates()
	if ret != nvml.SUCCESS {
		return domain.GPUMetrics{}, fmt.Errorf("failed to get utilization rates: %v", ret)
	}

	memory, ret := device.GetMemoryInfo()
	if ret != nvml.SUCCESS {
		return domain.GPUMetrics{}, fmt.Errorf("failed to get memory info: %v", ret)
	}

	return domain.GPUMetrics{
		GPUUsage:       float64(utilization.Gpu),
		GPUMemoryUsage: float64(memory.Used) / float64(memory.Total) * 100,
	}, nil
}
