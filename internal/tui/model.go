package tui

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/help"
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
	cpuUsageSparkline  *CPUUsageSparkline
	busiestCores       *BusiestCores
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
	help               help.Model
}

func InitialModel(collectors ...interface{}) (Model, error) {

	model := Model{
		cpuUsagePerCore: []float64{},
		cpuUsageTotal:   0,
		memoryUsage:     0,
		gpuUsage:        0,
		gpuMemoryUsage:  0,
		cpuUsageSparkline: NewCPUUsageSparkline(),
		cpuHeatmap: NewCPUHeatmap(),
		busiestCores: NewBusiestCores(),
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
		m.width = msg.Width
		m.height = msg.Height
		m.cpuHeatmap.Resize(m.width/2, 6)
		m.busiestCores.Resize(m.width/2, 6)
		m.cpuUsageSparkline.Resize(m.width/2, 3)
		log.Printf("Window size changed: %d x %d", msg.Width, msg.Height)
	case domain.CPUMemoryMetrics:
		m.cpuMemoryMetrics = msg
		m.cpuUsagePerCore = msg.CPUUsagePerCore
		m.cpuUsageTotal = msg.CPUUsageTotal
		m.memoryUsage = msg.MemoryUsage

		m.cpuUsageSparkline.Update(msg)
		m.cpuHeatmap.Update(msg)
		m.busiestCores.Update(msg)

		cmd = listenForMetrics(m.cpuMemoryCollector.Metrics())

	case domain.GPUMetrics:
		m.gpuUsage = msg.GPUUsage
		m.gpuMemoryUsage = msg.GPUMemoryUsage
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

	leftColumnStyle := lipgloss.NewStyle().
		Width(m.width / 2).
		Height(m.height - 2)

	rightColumnStyle := leftColumnStyle

	// Render components
	cpuUsageView := m.cpuUsageSparkline.View()
	cpuHeatmapView := m.cpuHeatmap.View()
	busiestCoresView := m.busiestCores.View()

	// Combine components
	leftColumn := lipgloss.JoinVertical(
		lipgloss.Left,
		cpuUsageView,
		lipgloss.PlaceHorizontal(m.width/2, lipgloss.Center, cpuHeatmapView),
		busiestCoresView,
	)

	rightColumn := lipgloss.JoinVertical(
		lipgloss.Left,
		fmt.Sprintf("Memory Usage: %.2f%%", m.memoryUsage),
		fmt.Sprintf("GPU Usage: %.2f%%", m.gpuUsage),
		fmt.Sprintf("GPU Memory Usage: %.2f%%", m.gpuMemoryUsage),
		"\nPress q to quit",
	)

	// Combine columns
	content := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftColumnStyle.Render(leftColumn),
		rightColumnStyle.Render(rightColumn),
	)

	// Render final view
	return containerStyle.Render(content)
}

func listenForMetrics[T any](metrics <-chan T) tea.Cmd {
	return func() tea.Msg {
		return <-metrics
	}
}
