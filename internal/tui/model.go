package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jonsampson/mim/internal/domain"
)

// metricsCollector is a private interface that defines the behavior we expect from a metrics collector
type metricsCollector[T any] interface {
	Start()
	Stop()
	Metrics() <-chan T
}

type Model struct {
	cpuHeatmap         *CPUHeatmap
	cpuMemoryMetrics   domain.CPUMemoryMetrics
	cpuMemoryCollector metricsCollector[domain.CPUMemoryMetrics]
	gpuCollector       metricsCollector[domain.GPUMetrics]
	cpuUsagePerCore    []float64
	cpuUsageTotal      float64
	memoryUsage        float64
	gpuUsage           float64
	gpuMemoryUsage     float64
}

func InitialModel(collectors ...interface{}) (Model, error) {
	model := Model{
		cpuUsagePerCore: []float64{},
		cpuUsageTotal:   0,
		memoryUsage:     0,
		gpuUsage:        0,
		gpuMemoryUsage:  0,
	}

	collectorInitialized := false

	for _, c := range collectors {
		switch collector := c.(type) {
		case metricsCollector[domain.CPUMemoryMetrics]:
			model.cpuMemoryCollector = collector
			collector.Start()
			collectorInitialized = true
		case metricsCollector[domain.GPUMetrics]:
			model.gpuCollector = collector
			collector.Start()
			collectorInitialized = true
		default:
			fmt.Printf("Unknown collector type: %T\n", c)
		}
	}

	if !collectorInitialized {
		return Model{}, fmt.Errorf("no valid collectors provided")
	}

	return model, nil
}

func (m Model) Init() tea.Cmd {
	var cmds []tea.Cmd

	if m.cpuMemoryCollector != nil {
		cmds = append(cmds, listenForMetrics(m.cpuMemoryCollector.Metrics()))
	}

	if m.gpuCollector != nil {
		cmds = append(cmds, listenForMetrics(m.gpuCollector.Metrics()))
	}

	return tea.Batch(cmds...)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.cpuMemoryCollector.Stop()
			if m.gpuCollector != nil {
				m.gpuCollector.Stop()
			}
			return m, tea.Quit
		}

	case domain.CPUMemoryMetrics:
		m.cpuMemoryMetrics = msg
		m.cpuUsagePerCore = msg.CPUUsagePerCore
		m.cpuUsageTotal = msg.CPUUsageTotal
		m.memoryUsage = msg.MemoryUsage

		// Initialize or update the CPU heatmap
		if m.cpuHeatmap == nil {
			m.cpuHeatmap = NewCPUHeatmap(len(m.cpuUsagePerCore))
		}
		m.cpuHeatmap.Update(msg)
		cmd = listenForMetrics(m.cpuMemoryCollector.Metrics())

	case domain.GPUMetrics:
		m.gpuUsage = msg.GPUUsage
		m.gpuMemoryUsage = msg.GPUMemoryUsage
		cmd = listenForMetrics(m.gpuCollector.Metrics())
	}

	return m, cmd
}

func (m Model) View() string {
	// Combine views of all components
	cpuHeatmapView := "CPU Heatmap initializing..."
	if m.cpuHeatmap != nil {
		cpuHeatmapView = m.cpuHeatmap.View()
	}
	return lipgloss.JoinVertical(
		lipgloss.Left,
		"CPU Usage",
		cpuHeatmapView,
		fmt.Sprintf("Total CPU Usage: %.2f%%", m.cpuUsageTotal),
		fmt.Sprintf("Memory Usage: %.2f%%", m.memoryUsage),
		fmt.Sprintf("GPU Usage: %.2f%%", m.gpuUsage),
		fmt.Sprintf("GPU Memory Usage: %.2f%%", m.gpuMemoryUsage),
		"\nPress q to quit",
	)
}

func listenForMetrics[T any](metrics <-chan T) tea.Cmd {
	return func() tea.Msg {
		return <-metrics
	}
}
