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

	processUtilizationList, err := device.GetProcessUtilization(1000000) // 1 second
	if err != nvml.SUCCESS {
		return domain.GPUMetrics{}, fmt.Errorf("failed to get process utilization info: %v", err)
	}

	graphicsRunningProcesses, err := device.GetGraphicsRunningProcesses()
	if err != nvml.SUCCESS {
		return domain.GPUMetrics{}, fmt.Errorf("failed to get graphics running processes: %v", err)
	}

	processInfo := make(map[uint32]domain.GPUProcessInfo)

	for _, process := range processUtilizationList {
		processInfo[process.Pid] = domain.GPUProcessInfo{
			Pid:    process.Pid,
			SmUtil: process.SmUtil,
		}
	}

	for _, process := range graphicsRunningProcesses {
		if info, exists := processInfo[process.Pid]; exists {
			info.UsedGpuMemory = process.UsedGpuMemory
			processInfo[process.Pid] = info
		} else {
			processInfo[process.Pid] = domain.GPUProcessInfo{
				Pid:           process.Pid,
				UsedGpuMemory: process.UsedGpuMemory,
			}
		}
	}

	processes := make([]domain.GPUProcessInfo, 0, len(processInfo))
	for _, info := range processInfo {
		processes = append(processes, info)
	}

	return domain.GPUMetrics{
		GPUUsage:       float64(utilization.Gpu),
		GPUMemoryUsage: float64(memory.Used) / float64(memory.Total) * 100,
		Processes:      processes,
	}, nil
}
