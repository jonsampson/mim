package tui

import (
	"math"

	"github.com/NimbleMarkets/ntcharts/heatmap"
	"github.com/jonsampson/mim/internal/domain"
)

type CPUHeatmap struct {
	hm              *heatmap.Model
	squareDimension int
	xOffset         int
	yOffset         int
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

func (c *CPUHeatmap) Resize(width, height int) {
	// Calculate the position to center the heatmap
	c.xOffset = (width - c.squareDimension) / 2
	c.yOffset = 0
}

func (c *CPUHeatmap) View() string {
	if c.hm == nil {
		return ""
	}
	// log.Printf("CPUHeatmap View: %v", c.hm)
	return c.hm.View()
	// view := c.hm.View()
	// lines := strings.Split(view, "\n")

	// // Add vertical padding
	// for range c.yOffset {
	// 	lines = append([]string{""}, lines...)
	// }

	// // Add horizontal padding
	// paddedLines := make([]string, len(lines))
	// for i, line := range lines {
	// 	paddedLines[i] = strings.Repeat(" ", c.xOffset) + line
	// }

	// return strings.Join(paddedLines, "\n")
}
