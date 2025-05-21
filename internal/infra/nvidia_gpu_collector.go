package infra

import (
	"fmt"
	"os"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/jonsampson/mim/internal/domain"
	"github.com/shirou/gopsutil/v3/process"
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

	type result struct {
		value any
		err   error
	}

	utilizationChan := make(chan result)
	memoryChan := make(chan result)
	processesChan := make(chan result)

	// Collect GPU utilization
	go func() {
		utilization, ret := device.GetUtilizationRates()
		if ret != nvml.SUCCESS {
			utilizationChan <- result{nil, fmt.Errorf("failed to get utilization rates: %v", ret)}
			return
		}
		utilizationChan <- result{float64(utilization.Gpu), nil}
	}()

	// Collect GPU memory usage
	go func() {
		memory, ret := device.GetMemoryInfo()
		if ret != nvml.SUCCESS {
			memoryChan <- result{nil, fmt.Errorf("failed to get memory info: %v", ret)}
			return
		}
		memoryChan <- result{float64(memory.Used) / float64(memory.Total) * 100, nil}
	}()

	// Collect process information
	go func() {
		processUtilizationList, ret := device.GetProcessUtilization(1000000) // 1 second
		if ret != nvml.SUCCESS && ret != nvml.ERROR_NOT_FOUND {
			processesChan <- result{nil, fmt.Errorf("failed to get process utilization info: %v", ret)}
			return
		}

		graphicsRunningProcesses, ret := device.GetGraphicsRunningProcesses()
		if ret != nvml.SUCCESS && ret != nvml.ERROR_NOT_FOUND {
			processesChan <- result{nil, fmt.Errorf("failed to get graphics running processes: %v", ret)}
			return
		}

		processInfo := make(map[uint32]domain.GPUProcessInfo)

		for _, process := range processUtilizationList {
			processInfo[process.Pid] = domain.GPUProcessInfo{
				Pid:    process.Pid,
				SmUtil: process.SmUtil,
			}
		}

		for _, process := range graphicsRunningProcesses {
			info, exists := processInfo[process.Pid]
			if exists {
				info.UsedGpuMemory = process.UsedGpuMemory
			} else {
				info = domain.GPUProcessInfo{
					Pid:           process.Pid,
					UsedGpuMemory: process.UsedGpuMemory,
				}
			}

			processInfo[process.Pid] = info
		}

		processes := make([]domain.GPUProcessInfo, 0, len(processInfo))
		for _, info := range processInfo {
			proc, err := process.NewProcess(int32(info.Pid))
			if err != nil {
				// Process might have terminated, log or handle as needed
				// For now, we'll skip adding user info if process lookup fails
				processes = append(processes, info)
				continue
			}
			username, err := proc.Username()
			if err != nil {
				// Handle error getting username, e.g., process terminated
				// For now, set to empty or a placeholder
				username = ""
				// Check if the error is os.ErrNotExist, if so, the process likely terminated
				// In other cases, you might want to log the error or handle it differently
				if !os.IsNotExist(err) {
					fmt.Printf("Error getting username for PID %d: %v\n", info.Pid, err)
				}
			}
			info.User = username
			processes = append(processes, info)
		}

		processesChan <- result{processes, nil}
	}()

	// Collect results and handle potential errors
	metrics := domain.GPUMetrics{}
	for range 3 {
		select {
		case r := <-utilizationChan:
			if r.err != nil {
				return domain.GPUMetrics{}, r.err
			}
			metrics.GPUUsage = r.value.(float64)
		case r := <-memoryChan:
			if r.err != nil {
				return domain.GPUMetrics{}, r.err
			}
			metrics.GPUMemoryUsage = r.value.(float64)
		case r := <-processesChan:
			if r.err != nil {
				return domain.GPUMetrics{}, r.err
			}
			metrics.Processes = r.value.([]domain.GPUProcessInfo)
		}
	}

	return metrics, nil
}
