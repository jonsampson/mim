package infra

import (
	"errors"
	"fmt"
	"os"
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
	type result struct {
		value any
		err   error
	}

	perCoreChan := make(chan result)
	totalChan := make(chan result)
	memChan := make(chan result)
	processesChan := make(chan result)

	// Collect per-core CPU usage
	go func() {
		cpuUsagePerCore, err := cpu.Percent(time.Second, true)
		perCoreChan <- result{cpuUsagePerCore, err}
	}()

	// Collect total CPU usage
	go func() {
		cpuUsageTotal, err := cpu.Percent(time.Second, false)
		if err != nil {
			totalChan <- result{nil, err}
		} else {
			totalChan <- result{cpuUsageTotal[0], nil}
		}
	}()

	// Get memory stats
	go func() {
		memStat, err := mem.VirtualMemory()
		if err != nil {
			memChan <- result{nil, err}
		} else {
			memChan <- result{memStat.UsedPercent, nil}
		}
	}()

	// Get process information
	go func() {
		processes, err := process.Processes()
		if err != nil {
			processesChan <- result{nil, err}
			return
		}

		processInfos := make([]domain.CPUProcessInfo, 0, len(processes))
		for _, p := range processes {
			pid := p.Pid
			cpuPercent, err := p.CPUPercent()
			if err != nil {
				if !errors.Is(err, os.ErrNotExist) {
					processesChan <- result{nil, fmt.Errorf("error getting CPU percent for PID %d: %w", pid, err)}
					return
				}
				continue
			}
			memPercent, err := p.MemoryPercent()
			if err != nil {
				if !errors.Is(err, os.ErrNotExist) {
					processesChan <- result{nil, fmt.Errorf("error getting memory percent for PID %d: %w", pid, err)}
					return
				}
				continue
			}
			name, err := p.Name()
			if err != nil {
				if !errors.Is(err, os.ErrNotExist) {
					processesChan <- result{nil, fmt.Errorf("error getting name for PID %d: %w", pid, err)}
					return
				}
				continue
			}

			processInfos = append(processInfos, domain.CPUProcessInfo{
				Pid:           uint32(pid),
				CPUPercent:    cpuPercent,
				MemoryPercent: float64(memPercent),
				Command:       name,
			})
		}
		processesChan <- result{processInfos, nil}
	}()

	// Collect results and handle potential errors
	metrics := domain.CPUMemoryMetrics{}
	var err error

	for range 4 {
		select {
		case r := <-perCoreChan:
			if r.err != nil {
				return domain.CPUMemoryMetrics{}, r.err
			}
			metrics.CPUUsagePerCore = r.value.([]float64)
		case r := <-totalChan:
			if r.err != nil {
				return domain.CPUMemoryMetrics{}, r.err
			}
			metrics.CPUUsageTotal = r.value.(float64)
		case r := <-memChan:
			if r.err != nil {
				return domain.CPUMemoryMetrics{}, r.err
			}
			metrics.MemoryUsage = r.value.(float64)
		case r := <-processesChan:
			if r.err != nil {
				return domain.CPUMemoryMetrics{}, r.err
			}
			metrics.Processes = r.value.([]domain.CPUProcessInfo)
		}
	}

	// Return the collected metrics
	return metrics, err
}
