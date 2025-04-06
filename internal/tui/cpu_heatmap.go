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

func NewCPUHeatmap() *CPUHeatmap {
	return &CPUHeatmap{}
}

func (c *CPUHeatmap) Update(metrics domain.CPUMemoryMetrics) {
	c.squareDimension = int(math.Ceil(math.Sqrt(float64(len(metrics.CPUUsagePerCore)))))
	heatMap := heatmap.New(c.squareDimension+1,
		c.squareDimension+1,
		heatmap.WithValueRange(0, 100),
	)
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
	heatMap.PushAllMatrixRow(matrix)
	heatMap.Draw()
	c.hm = &heatMap
}

func (c *CPUHeatmap) View() string {
	if c.hm == nil {
		return ""
	}
	return c.hm.View()
}
