package tui

import (
    "fmt"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/jonsampson/mim/internal/infra"
)

// metricsCollector is a private interface that defines the behavior we expect from a metrics collector
type metricsCollector interface {
    Start()
    Stop()
    Metrics() <-chan infra.SystemMetrics
}
type Model struct {
    metricsCollector metricsCollector // Changed from *infra.MetricsCollector to the interface
    cpuUsage         float64
    memoryUsage      float64
    gpuUsage         float64
}

func InitialModel() Model {
    mc := infra.NewMetricsCollector()
    mc.Start()
    return Model{
        metricsCollector: mc,
        cpuUsage:         0,
        memoryUsage:      0,
        gpuUsage:         0,
    }
}

func (m Model) Init() tea.Cmd {
    return listenForMetrics(m.metricsCollector.Metrics())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "q", "ctrl+c":
            m.metricsCollector.Stop()
            return m, tea.Quit
        }

    case infra.SystemMetrics:
        m.cpuUsage = msg.CPUUsage
        m.memoryUsage = msg.MemoryUsage
        return m, listenForMetrics(m.metricsCollector.Metrics())
    }

    return m, nil
}

func (m Model) View() string {
    return fmt.Sprintf("CPU Usage: %.2f%%\nMemory Usage: %.2f%%\nGPU Usage: %.2f%%\n\nPress q to quit", m.cpuUsage, m.memoryUsage, m.gpuUsage)
}

func listenForMetrics(metrics <-chan infra.SystemMetrics) tea.Cmd {
    return func() tea.Msg {
        return <-metrics
    }
}
