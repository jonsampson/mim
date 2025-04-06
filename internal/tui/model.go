package tui

import (
	"fmt"
	"log"

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
	cpuGPUUsageGraph   *CPUGPUUsageGraph
	memoryUsageGraph   *MemoryUsageGraph
	cpuMemoryMetrics   domain.CPUMemoryMetrics
	cpuMemoryCollector metricsCollector[domain.CPUMemoryMetrics]
	gpuCollector       metricsCollector[domain.GPUMetrics]
	cpuUsagePerCore    []float64
	cpuUsageTotal      float64
	memoryUsage        float64
	gpuUsage           float64
	gpuMemoryUsage     float64
	width              int
	height             int
	cpuCombinedView    *CPUCombinedView
}

func InitialModel(collectors ...interface{}) (Model, error) {

	model := Model{
		cpuUsagePerCore:  []float64{},
		cpuUsageTotal:    0,
		memoryUsage:      0,
		gpuUsage:         0,
		gpuMemoryUsage:   0,
		cpuGPUUsageGraph: NewCPUGPUUsageGraph(),
		memoryUsageGraph: NewMemoryUsageGraph(),
		cpuCombinedView:  NewCPUCombinedView(),
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

	case tea.WindowSizeMsg:
		log.Printf("Window size changed: %d x %d", msg.Width, msg.Height)
		m.width = msg.Width
		m.height = msg.Height
		m.cpuCombinedView.Resize(m.width-5, m.height)
		m.cpuGPUUsageGraph.Resize(m.width-5, 10)
		m.memoryUsageGraph.Resize(m.width-5, 10)
	case domain.CPUMemoryMetrics:
		m.cpuMemoryMetrics = msg
		m.cpuUsagePerCore = msg.CPUUsagePerCore
		m.cpuUsageTotal = msg.CPUUsageTotal
		m.memoryUsage = msg.MemoryUsage

		m.cpuCombinedView.Update(msg)
		m.cpuGPUUsageGraph.UpdateCPU(msg)
		m.memoryUsageGraph.UpdateSystemMemory(msg)

		cmd = listenForMetrics(m.cpuMemoryCollector.Metrics())

	case domain.GPUMetrics:
		m.gpuUsage = msg.GPUUsage
		m.gpuMemoryUsage = msg.GPUMemoryUsage
		m.cpuGPUUsageGraph.UpdateGPU(msg)
		m.memoryUsageGraph.UpdateGPUMemory(msg)

		cmd = listenForMetrics(m.gpuCollector.Metrics())
	}

	return m, cmd
}

func (m Model) View() string {
	// Define styles for layout
	containerStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Padding(1)

	// Render components
	cpuSection := m.cpuCombinedView.View()

	// Combine columns
	content := lipgloss.JoinVertical(
		lipgloss.Top,
		m.cpuGPUUsageGraph.View(),
		fmt.Sprintf("    CPU Usage: %.2f%%   GPU Usage: %.2f%%", m.cpuUsageTotal, m.gpuUsage),
		lipgloss.NewStyle().Render(""),
		cpuSection,
		m.memoryUsageGraph.View(),
		fmt.Sprintf("    Memory Usage: %.2f%%   GPU Memory Usage: %.2f%%", m.memoryUsage, m.gpuMemoryUsage),
		"\nPress q to quit",
	)

	// Render final view
	return containerStyle.Render(content)
}

func listenForMetrics[T any](metrics <-chan T) tea.Cmd {
	return func() tea.Msg {
		return <-metrics
	}
}
