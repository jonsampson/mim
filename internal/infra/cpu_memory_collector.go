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
	cpuUsage, err := cpu.Percent(time.Second, true)
	if err != nil {
		return domain.CPUMemoryMetrics{}, err
	}

	memStat, err := mem.VirtualMemory()
	if err != nil {
		return domain.CPUMemoryMetrics{}, err
	}

	return domain.CPUMemoryMetrics{
		CPUUsage:    cpuUsage,
		MemoryUsage: memStat.UsedPercent,
	}, nil
}
