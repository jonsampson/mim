package tui

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/viewport"
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
	gpuMetrics         domain.GPUMetrics
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
	processMonitor     *ProcessMonitor
	viewport           viewport.Model
}

func InitialModel(collectors ...any) (Model, error) {
	model := Model{
		cpuUsagePerCore:  []float64{},
		cpuUsageTotal:    0,
		memoryUsage:      0,
		gpuUsage:         0,
		gpuMemoryUsage:   0,
		cpuGPUUsageGraph: NewCPUGPUUsageGraph(),
		memoryUsageGraph: NewMemoryUsageGraph(),
		cpuCombinedView:  NewCPUCombinedView(),
		processMonitor:   NewProcessMonitor(80), // Initialize with a default width
		width:            80,                    // Set a default width
		height:           24,                    // Set a default height
		viewport:         viewport.New(80, 24),
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

	// Return a command to get the initial window size
	cmds = append(cmds, tea.EnterAltScreen)

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
		case "up", "k":
			m.viewport.LineUp(1)
		case "down", "j":
			m.viewport.LineDown(1)
		case "pgup":
			m.viewport.HalfViewUp()
		case "pgdown":
			m.viewport.HalfViewDown()
		case "home":
			m.viewport.GotoTop()
		case "end":
			m.viewport.GotoBottom()
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		log.Printf("Window size changed: %d x %d", msg.Width, msg.Height)
		m.width = msg.Width
		m.height = msg.Height
		m.processMonitor.Resize(m.width) // Add this method to ProcessMonitor
		m.cpuCombinedView.Resize(m.width-5, m.height)
		m.cpuGPUUsageGraph.Resize(m.width-5, 10)
		m.memoryUsageGraph.Resize(m.width-5, 10)
		m.viewport.Height = m.height
		m.viewport.Width = m.width
	case domain.CPUMemoryMetrics:
		m.cpuMemoryMetrics = msg
		m.cpuUsagePerCore = msg.CPUUsagePerCore
		m.cpuUsageTotal = msg.CPUUsageTotal
		m.memoryUsage = msg.MemoryUsage

		m.cpuCombinedView.Update(msg)
		m.cpuGPUUsageGraph.Update(msg)
		m.memoryUsageGraph.Update(msg)

		m.processMonitor.UpdateProcesses(m.cpuMemoryMetrics.Processes, m.gpuMetrics.Processes)

		m.viewport.SetContent(m.renderContent())
		cmd = listenForMetrics(m.cpuMemoryCollector.Metrics())

	case domain.GPUMetrics:
		m.gpuMetrics = msg
		m.gpuUsage = msg.GPUUsage
		m.gpuMemoryUsage = msg.GPUMemoryUsage
		m.cpuGPUUsageGraph.Update(msg)
		m.memoryUsageGraph.Update(msg)

		m.processMonitor.UpdateProcesses(m.cpuMemoryMetrics.Processes, m.gpuMetrics.Processes)

		m.viewport.SetContent(m.renderContent())
		cmd = listenForMetrics(m.gpuCollector.Metrics())
	}

	return m, cmd
}

func (m Model) View() string {
	return fmt.Sprintf("%s\n%s", m.viewport.View(), m.statusBarView())
}

// Add a new method to render the content
func (m Model) renderContent() string {
	// Render components
	cpuSection := m.cpuCombinedView.View()

	// Combine columns
	content := lipgloss.JoinVertical(
		lipgloss.Top, "\n", // add spacing for viewport
		m.cpuGPUUsageGraph.View(),
		fmt.Sprintf("    CPU Usage: %.2f%%   GPU Usage: %.2f%%", m.cpuUsageTotal, m.gpuUsage),
		lipgloss.NewStyle().Margin(0).Render(""),
		cpuSection,
		m.memoryUsageGraph.View(),
		fmt.Sprintf("    Memory Usage: %.2f%%   GPU Memory Usage: %.2f%%", m.memoryUsage, m.gpuMemoryUsage),
		m.processMonitor.View(),
	)

	return content
}

func listenForMetrics[T any](metrics <-chan T) tea.Cmd {
	return func() tea.Msg {
		return <-metrics
	}
}

func (m Model) statusBarView() string {
	return fmt.Sprintf("Press q to quit | Scroll: ↑/↓ or mouse | %3.f%%", m.viewport.ScrollPercent()*100)
}
