package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jonsampson/mim/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestModelUpdate(t *testing.T) {
	mockCPUMemCollector := new(MockMetricsCollector[domain.CPUMemoryMetrics])
	mockGPUCollector := new(MockMetricsCollector[domain.GPUMetrics])

	model := Model{
		cpuMemoryCollector: mockCPUMemCollector,
		gpuCollector:       mockGPUCollector,
	}

	t.Run("Update with CPU/Memory metrics", func(t *testing.T) {
		cpuMemMetrics := domain.CPUMemoryMetrics{
			CPUUsage:    []float64{50.0, 60.0},
			MemoryUsage: 70.0,
		}

		updatedModel, cmd := model.Update(cpuMemMetrics)
		assert.Equal(t, cpuMemMetrics.CPUUsage, updatedModel.(Model).cpuUsage)
		assert.Equal(t, cpuMemMetrics.MemoryUsage, updatedModel.(Model).memoryUsage)
		assert.Nil(t, cmd)
	})

	t.Run("Update with GPU metrics", func(t *testing.T) {
		gpuMetrics := domain.GPUMetrics{
			GPUUsage:       80.0,
			GPUMemoryUsage: 90.0,
		}

		updatedModel, cmd := model.Update(gpuMetrics)
		assert.Equal(t, gpuMetrics.GPUUsage, updatedModel.(Model).gpuUsage)
		assert.Equal(t, gpuMetrics.GPUMemoryUsage, updatedModel.(Model).gpuMemoryUsage)
		assert.Nil(t, cmd)
	})

	t.Run("Update with quit message", func(t *testing.T) {
		mockCPUMemCollector.On("Stop").Once()
		mockGPUCollector.On("Stop").Once()

		_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
		assert.NotNil(t, cmd)
		expectedResult := tea.Quit()
		actualResult := cmd()
		assert.Equal(t, expectedResult, actualResult)

		mockCPUMemCollector.AssertExpectations(t)
		mockGPUCollector.AssertExpectations(t)
	})
}

func TestModelView(t *testing.T) {
	model := Model{
		cpuUsage:       []float64{50.0, 60.0},
		memoryUsage:    70.0,
		gpuUsage:       80.0,
		gpuMemoryUsage: 90.0,
	}

	view := model.View()
	assert.Contains(t, view, "CPU Usage: Core 0: 50.00% Core 1: 60.00%")
	assert.Contains(t, view, "Memory Usage: 70.00%")
	assert.Contains(t, view, "GPU Usage: 80.00%")
	assert.Contains(t, view, "GPU Memory Usage: 90.00%")
	assert.Contains(t, view, "Press q to quit")
}
