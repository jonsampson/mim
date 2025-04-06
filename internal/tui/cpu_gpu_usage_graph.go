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

func (g *CPUGPUUsageGraph) UpdateCPU(cpuMetrics domain.CPUMemoryMetrics) {
	g.slc.PushDataSet(cpuDataSet, cpuMetrics.CPUUsageTotal)
	g.slc.DrawAll()
}
func (g *CPUGPUUsageGraph) UpdateGPU(gpuMetrics domain.GPUMetrics) {
	g.slc.PushDataSet(gpuDataSet, gpuMetrics.GPUUsage)
	g.slc.DrawAll()
}

func (g *CPUGPUUsageGraph) View() string {
	return g.slc.View()
}

func (g *CPUGPUUsageGraph) Resize(width, height int) {
	g.slc.Resize(width, height)
}