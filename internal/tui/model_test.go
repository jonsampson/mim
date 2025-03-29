package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jonsampson/mim/internal/infra"
	"github.com/stretchr/testify/assert"
)

func TestInitialModel(t *testing.T) {
	model := InitialModel()
	assert.NotNil(t, model.metricsCollector)
	assert.Equal(t, 0.0, model.cpuUsage)
	assert.Equal(t, 0.0, model.memoryUsage)
	assert.Equal(t, 0.0, model.gpuUsage)
}

func TestModelUpdate(t *testing.T) {
    mockCollector := &MockMetricsCollector{
        metricsChan: make(chan infra.SystemMetrics, 1),
    }
    mockCollector.On("Start").Return()
    mockCollector.On("Stop").Return()

    model := Model{
        metricsCollector: mockCollector,
    }

    t.Run("Quit on 'q' key press", func(t *testing.T) {
        updatedModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
        assert.Equal(t, model, updatedModel) // Check that the model hasn't changed
        assert.Equal(t, tea.Quit, cmd())
        mockCollector.AssertCalled(t, "Stop")
    })

    t.Run("Update metrics", func(t *testing.T) {
        metrics := infra.SystemMetrics{
            CPUUsage:    50.0,
            MemoryUsage: 60.0,
        }
        mockCollector.metricsChan <- metrics

        newModel, cmd := model.Update(<-mockCollector.metricsChan)
        updatedModel, ok := newModel.(Model)
        assert.True(t, ok)
        assert.Equal(t, 50.0, updatedModel.cpuUsage)
        assert.Equal(t, 60.0, updatedModel.memoryUsage)
        assert.NotNil(t, cmd)
    })
}

func TestModelView(t *testing.T) {
	model := Model{
		cpuUsage:    30.5,
		memoryUsage: 45.7,
		gpuUsage:    0.0,
	}

	view := model.View()
	assert.Contains(t, view, "CPU Usage: 30.50%")
	assert.Contains(t, view, "Memory Usage: 45.70%")
	assert.Contains(t, view, "GPU Usage: 0.00%")
	assert.Contains(t, view, "Press q to quit")
}
