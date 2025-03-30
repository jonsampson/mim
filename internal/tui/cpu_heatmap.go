package tui

import (
	"math"

	"github.com/NimbleMarkets/ntcharts/heatmap"
	"github.com/jonsampson/mim/internal/domain"
)

type CPUHeatmap struct {
	hm              *heatmap.Model
	squareDimension int
}

func NewCPUHeatmap(cores int) *CPUHeatmap {
	squareDimension := int(math.Ceil(math.Sqrt(float64(cores))))
	hm := heatmap.New(squareDimension + 1,
		squareDimension + 1,
		heatmap.WithValueRange(0, 100),
	)
	return &CPUHeatmap{hm: &hm, squareDimension: squareDimension}
}

func (c *CPUHeatmap) Update(metrics domain.CPUMemoryMetrics) {
	c.hm.Clear()
	matrix := make([][]float64, c.squareDimension)
	core := 0
	for i := range matrix {
        matrix[i] = make([]float64, c.squareDimension)
		for j := range matrix[i] {
			if core >= len(metrics.CPUUsagePerCore) {
                break
            }
            matrix[i][j] = metrics.CPUUsagePerCore[core]
			core++
        }
    }
	c.hm.PushAllMatrixRow(matrix)
	c.hm.Draw()
}

func (c *CPUHeatmap) View() string {
	return c.hm.View()
}
