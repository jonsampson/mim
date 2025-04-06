package tui

import (
    "github.com/NimbleMarkets/ntcharts/canvas/runes"
    "github.com/NimbleMarkets/ntcharts/linechart/streamlinechart"
    "github.com/charmbracelet/lipgloss"
    "github.com/jonsampson/mim/internal/domain"
)

const (
    systemMemoryDataSet = "System"
    gpuMemoryDataSet    = "GPU"
)

var graphLineStyleSystem = lipgloss.NewStyle().
    Foreground(lipgloss.Color("5")) // magenta

var graphLineStyleGPUMem = lipgloss.NewStyle().
    Foreground(lipgloss.Color("11")) // yellow

var axisStyleMem = lipgloss.NewStyle().
    Foreground(lipgloss.Color("3")) // yellow

var labelStyleMem = lipgloss.NewStyle().
    Foreground(lipgloss.Color("6")) // cyan

type MemoryUsageGraph struct {
    slc streamlinechart.Model
}

func NewMemoryUsageGraph() *MemoryUsageGraph {
    slc := streamlinechart.New(10, 10,
        streamlinechart.WithYRange(0, 100),
        streamlinechart.WithAxesStyles(axisStyleMem, labelStyleMem),
        streamlinechart.WithStyles(runes.ThinLineStyle, graphLineStyleSystem),
        streamlinechart.WithDataSetStyles(systemMemoryDataSet, runes.ThinLineStyle, graphLineStyleSystem),
        streamlinechart.WithDataSetStyles(gpuMemoryDataSet, runes.ThinLineStyle, graphLineStyleGPUMem),
    )

    slc.DrawXYAxisAndLabel()

    return &MemoryUsageGraph{
        slc: slc,
    }
}

func (g *MemoryUsageGraph) Update(msg interface{}) {
    switch msg := msg.(type) {
    case domain.CPUMemoryMetrics:
        g.updateSystemMemory(msg)
    case domain.GPUMetrics:
        g.updateGPUMemory(msg)
    }
}

func (g *MemoryUsageGraph) updateSystemMemory(memoryMetrics domain.CPUMemoryMetrics) {
    g.slc.PushDataSet(systemMemoryDataSet, memoryMetrics.MemoryUsage)
}

func (g *MemoryUsageGraph) updateGPUMemory(gpuMetrics domain.GPUMetrics) {
    g.slc.PushDataSet(gpuMemoryDataSet, gpuMetrics.GPUMemoryUsage)
}

func (g *MemoryUsageGraph) View() string {
    g.slc.DrawAll()
    return g.slc.View()
}

func (g *MemoryUsageGraph) Resize(width, height int) {
    g.slc.Resize(width, height)
}