package tui

import (
	"github.com/NimbleMarkets/ntcharts/canvas/runes"
	"github.com/NimbleMarkets/ntcharts/linechart/streamlinechart"
	"github.com/charmbracelet/lipgloss"
	"github.com/jonsampson/mim/internal/domain"
)

const (
	cpuDataSet = "CPU"
	gpuDataSet = "GPU"
)

var graphLineStyleCPU = lipgloss.NewStyle().
	Foreground(lipgloss.Color("4")) // blue

var graphLineStyleGPU = lipgloss.NewStyle().
	Foreground(lipgloss.Color("10")) // green

var axisStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("3")) // yellow

var labelStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("6")) // cyan

type CPUGPUUsageGraph struct {
	slc streamlinechart.Model
}

func NewCPUGPUUsageGraph() *CPUGPUUsageGraph {
	slc := streamlinechart.New(10, 10,
		streamlinechart.WithYRange(0, 100),
		streamlinechart.WithAxesStyles(axisStyle, labelStyle),
		streamlinechart.WithStyles(runes.ThinLineStyle, graphLineStyleCPU),
		streamlinechart.WithDataSetStyles(gpuDataSet, runes.ThinLineStyle, graphLineStyleGPU),
		streamlinechart.WithDataSetStyles(cpuDataSet, runes.ThinLineStyle, graphLineStyleCPU),
	)

	slc.DrawXYAxisAndLabel()

	return &CPUGPUUsageGraph{
		slc: slc,
	}
}

func (g *CPUGPUUsageGraph) Update(msg interface{}) {
	switch msg := msg.(type) {
	case domain.CPUMemoryMetrics:
		g.updateCPU(msg)
	case domain.GPUMetrics:
		g.updateGPU(msg)
	}
}

func (g *CPUGPUUsageGraph) updateCPU(cpuMetrics domain.CPUMemoryMetrics) {
	g.slc.PushDataSet(cpuDataSet, cpuMetrics.CPUUsageTotal)
}

func (g *CPUGPUUsageGraph) updateGPU(gpuMetrics domain.GPUMetrics) {
	g.slc.PushDataSet(gpuDataSet, gpuMetrics.GPUUsage)
}

func (g *CPUGPUUsageGraph) View() string {
	g.slc.DrawAll()
	return g.slc.View()
}

func (g *CPUGPUUsageGraph) Resize(width, height int) {
	g.slc.Resize(width, height)
}