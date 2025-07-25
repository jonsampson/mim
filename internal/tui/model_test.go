package tui

import (
	"testing"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/jonsampson/mim/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestModelUpdate(t *testing.T) {
	mockCPUMemoryCollector := new(MockMetricsCollector[domain.CPUMemoryMetrics])
	mockGPUCollector := new(MockMetricsCollector[domain.GPUMetrics])

	// Set up expectations for Metrics() method
	cpuMemoryMetricsChan := make(chan domain.CPUMemoryMetrics, 1)
	gpuMetricsChan := make(chan domain.GPUMetrics, 1)
	mockCPUMemoryCollector.On("Metrics").Return(cpuMemoryMetricsChan)
	mockGPUCollector.On("Metrics").Return(gpuMetricsChan)

	model := Model{
		cpuMemoryCollector: mockCPUMemoryCollector,
		gpuCollector:       mockGPUCollector,
		cpuCombinedView:    NewCPUCombinedView(),
		cpuGPUUsageGraph:   NewCPUGPUUsageGraph(),
		memoryUsageGraph:   NewMemoryUsageGraph(),
		processMonitor:     NewProcessMonitor(80),
	}
	t.Run("CPU and Memory metrics update", func(t *testing.T) {
		cpuMemoryMetrics := domain.CPUMemoryMetrics{
			CPUUsagePerCore: []float64{10.0, 20.0},
			CPUUsageTotal:   15.0,
			MemoryUsage:     50.0,
		}

		updatedModel, cmd := model.Update(cpuMemoryMetrics)
		updatedModelTyped := updatedModel.(Model)

		assert.Equal(t, cpuMemoryMetrics.CPUUsagePerCore, updatedModelTyped.cpuUsagePerCore)
		assert.Equal(t, cpuMemoryMetrics.CPUUsageTotal, updatedModelTyped.cpuUsageTotal)
		assert.Equal(t, cpuMemoryMetrics.MemoryUsage, updatedModelTyped.memoryUsage)
		assert.NotNil(t, cmd)
	})

	t.Run("GPU metrics update", func(t *testing.T) {
		gpuMetrics := domain.GPUMetrics{
			GPUUsage:       70.0,
			GPUMemoryUsage: 80.0,
		}

		updatedModel, cmd := model.Update(gpuMetrics)
		updatedModelTyped := updatedModel.(Model)

		assert.Equal(t, gpuMetrics.GPUUsage, updatedModelTyped.gpuUsage)
		assert.Equal(t, gpuMetrics.GPUMemoryUsage, updatedModelTyped.gpuMemoryUsage)
		assert.NotNil(t, cmd)
	})

	// Verify that the expectations were met
	mockCPUMemoryCollector.AssertExpectations(t)
	mockGPUCollector.AssertExpectations(t)
}

func TestModelView(t *testing.T) {
    model := Model{
        cpuUsagePerCore: []float64{10.0, 20.0},
        cpuUsageTotal:   15.0,
        memoryUsage:     50.0,
        gpuUsage:        70.0,
        gpuMemoryUsage:  80.0,
        cpuCombinedView: NewCPUCombinedView(),
        cpuGPUUsageGraph: NewCPUGPUUsageGraph(),
        memoryUsageGraph: NewMemoryUsageGraph(),
        processMonitor:   NewProcessMonitor(80),
        viewport:         viewport.New(80, 50),
    }

    // Set up the viewport with the rendered content
    model.viewport.SetContent(model.renderContent())
    
    view := model.View()

    assert.Contains(t, view, "CPU Usage: 15.00%")
    assert.Contains(t, view, "Memory Usage: 50.00%")
    assert.Contains(t, view, "GPU Usage: 70.00%")
    assert.Contains(t, view, "GPU Memory Usage: 80.00%")
    assert.Contains(t, view, "Press q to quit")
}

func TestInitialModel(t *testing.T) {
	mockCPUMemoryCollector := new(MockMetricsCollector[domain.CPUMemoryMetrics])
	mockGPUCollector := new(MockMetricsCollector[domain.GPUMetrics])

	mockCPUMemoryCollector.On("Start").Return()
	mockGPUCollector.On("Start").Return()

	model, err := InitialModel(mockCPUMemoryCollector, mockGPUCollector)

	assert.NoError(t, err)
	assert.NotNil(t, model.cpuMemoryCollector)
	assert.NotNil(t, model.gpuCollector)

	mockCPUMemoryCollector.AssertCalled(t, "Start")
	mockGPUCollector.AssertCalled(t, "Start")
}

func TestInitialModelError(t *testing.T) {
	_, err := InitialModel()

	assert.Error(t, err)
	assert.Equal(t, "no valid collectors provided", err.Error())
}
