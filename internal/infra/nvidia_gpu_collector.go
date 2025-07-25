package infra

import (
	"fmt"
	"log"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/jonsampson/mim/internal/domain"
)

type NvidiaGPUCollector struct {
	*BaseCollector[domain.GPUMetrics]
	gpuCalculator *domain.GPUCalculator
	usernameCache *UsernameCache
}

func NewNvidiaGPUCollector() *NvidiaGPUCollector {
	collector := &NvidiaGPUCollector{
		gpuCalculator: domain.NewGPUCalculator(),
		usernameCache: NewUsernameCache(),
	}
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
		memoryPercent := c.gpuCalculator.CalculateMemoryPercent(memory.Used, memory.Total)
		memoryChan <- result{memoryPercent, nil}
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

		computeRunningProcesses, ret := device.GetComputeRunningProcesses()
		if ret != nvml.SUCCESS && ret != nvml.ERROR_NOT_FOUND {
			processesChan <- result{nil, fmt.Errorf("failed to get compute running processes: %v", ret)}
			return
		}

		processInfo := make(map[uint32]domain.GPUProcessInfo)

		for _, process := range processUtilizationList {
			processInfo[process.Pid] = domain.GPUProcessInfo{
				Pid:    process.Pid,
				SmUtil: process.SmUtil,
			}
		}

		memory, ret := device.GetMemoryInfo()
		if ret != nvml.SUCCESS {
			memoryChan <- result{nil, fmt.Errorf("failed to get memory info: %v", ret)}
			return
		}

		// Process both graphics and compute processes for memory usage
		allProcesses := append(graphicsRunningProcesses, computeRunningProcesses...)
		for _, process := range allProcesses {
			info, exists := processInfo[process.Pid]
			processMemoryPercent := c.gpuCalculator.CalculateProcessMemoryPercent(process.UsedGpuMemory, memory.Total)
			if exists {
				info.UsedGpuMemory = processMemoryPercent
			} else {
				info = domain.GPUProcessInfo{
					Pid:           process.Pid,
					UsedGpuMemory: processMemoryPercent,
				}
			}
			log.Printf("Processing GPU process: PID=%d, PctUsed=%f, UsedGpuMemory=%v, TotalMemory=%d", process.Pid, info.UsedGpuMemory, process.UsedGpuMemory, memory.Total)

			processInfo[process.Pid] = info
		}

		log.Printf("Total GPU processes found: %d", len(processInfo))
		processes := make([]domain.GPUProcessInfo, 0, len(processInfo))
		for _, info := range processInfo {
			// Get username using cache (handles all error cases with timeout protection)
			info.User = c.usernameCache.GetUsername(info.Pid)
			processes = append(processes, info)
		}

		log.Printf("Sending %d GPU processes to UI", len(processes))
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
