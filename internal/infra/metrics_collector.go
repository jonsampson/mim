package infra

import (
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type MetricsCollector struct {
	metrics chan SystemMetrics
	stop    chan struct{}
}

type SystemMetrics struct {
	CPUUsage    float64
	MemoryUsage float64
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics: make(chan SystemMetrics),
		stop:    make(chan struct{}),
	}
}

func (mc *MetricsCollector) Start() {
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				metrics, err := mc.collectMetrics()
				if err == nil {
					mc.metrics <- metrics
				}
			case <-mc.stop:
				return
			}
		}
	}()
}

func (mc *MetricsCollector) Stop() {
	close(mc.stop)
}

func (mc *MetricsCollector) Metrics() <-chan SystemMetrics {
	return mc.metrics
}

func (mc *MetricsCollector) collectMetrics() (SystemMetrics, error) {
	cpuUsage, err := mc.getCPUUsage()
	if err != nil {
		return SystemMetrics{}, err
	}

	memoryUsage, err := mc.getMemoryUsage()
	if err != nil {
		return SystemMetrics{}, err
	}

	return SystemMetrics{
		CPUUsage:    cpuUsage,
		MemoryUsage: memoryUsage,
	}, nil
}

func (mc *MetricsCollector) getCPUUsage() (float64, error) {
	percentage, err := cpu.Percent(time.Second, false)
	if err != nil {
		return 0, err
	}
	return percentage[0], nil
}

func (mc *MetricsCollector) getMemoryUsage() (float64, error) {
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return 0, err
	}
	return vmStat.UsedPercent, nil
}
