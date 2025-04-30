package infra

import (
	"time"

	"github.com/jonsampson/mim/internal/domain"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

type CPUMemoryCollector struct {
	*BaseCollector[domain.CPUMemoryMetrics]
}

func NewCPUMemoryCollector() *CPUMemoryCollector {
	collector := &CPUMemoryCollector{}
	collector.BaseCollector = NewBaseCollector(collector.getMetrics)
	return collector
}

func (c *CPUMemoryCollector) getMetrics() (domain.CPUMemoryMetrics, error) {
	// Create channels for results and errors
	perCoreChan := make(chan []float64)
	totalChan := make(chan float64)
	errChan := make(chan error, 2) // Buffer of 2 to avoid goroutine leak

	// Collect per-core CPU usage in a goroutine
	go func() {
		cpuUsagePerCore, err := cpu.Percent(time.Second, true)
		if err != nil {
			errChan <- err
			return
		}
		perCoreChan <- cpuUsagePerCore
	}()

	// Collect total CPU usage in a goroutine
	go func() {
		cpuUsageTotal, err := cpu.Percent(time.Second, false)
		if err != nil {
			errChan <- err
			return
		}
		totalChan <- cpuUsageTotal[0]
	}()

	// Collect results and handle potential errors
	var cpuUsagePerCore []float64
	var cpuUsageTotal float64
	for range 2 {
		select {
		case perCore := <-perCoreChan:
			cpuUsagePerCore = perCore
		case total := <-totalChan:
			cpuUsageTotal = total
		case err := <-errChan:
			return domain.CPUMemoryMetrics{}, err
		}
	}

	// Get memory stats
	memStat, err := mem.VirtualMemory()
	if err != nil {
		return domain.CPUMemoryMetrics{}, err
	}

	// Get process information
	processes, err := process.Processes()
	if err != nil {
		return domain.CPUMemoryMetrics{}, err
	}

	metrics := domain.CPUMemoryMetrics{
		CPUUsagePerCore: cpuUsagePerCore,
		CPUUsageTotal:   cpuUsageTotal,
		MemoryUsage:     memStat.UsedPercent,
		Processes:       make([]domain.CPUProcessInfo, 0, len(processes)),
	}

	for _, p := range processes {
		pid := p.Pid
		cpuPercent, _ := p.CPUPercent()
		memPercent, _ := p.MemoryPercent()
		name, _ := p.Name()

		metrics.Processes = append(metrics.Processes, domain.CPUProcessInfo{
			Pid:           uint32(pid),
			CPUPercent:    cpuPercent,
			MemoryPercent: float64(memPercent),
			Command:       name,
		})
	}

	// Return the collected metrics
	return metrics, nil
}
