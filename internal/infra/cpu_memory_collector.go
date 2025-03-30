package infra

import (
	"time"

	"github.com/jonsampson/mim/internal/domain"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
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
    // Get per-core CPU usage
    cpuUsagePerCore, err := cpu.Percent(time.Second, true)
    if err != nil {
        return domain.CPUMemoryMetrics{}, err
    }

    // Get total CPU usage
    cpuUsageTotal, err := cpu.Percent(time.Second, false)
    if err != nil {
        return domain.CPUMemoryMetrics{}, err
    }

    memStat, err := mem.VirtualMemory()
    if err != nil {
        return domain.CPUMemoryMetrics{}, err
    }

    return domain.CPUMemoryMetrics{
        CPUUsagePerCore: cpuUsagePerCore,
        CPUUsageTotal:   cpuUsageTotal[0], // cpuUsageTotal is a slice with one element
        MemoryUsage:     memStat.UsedPercent,
    }, nil
}
