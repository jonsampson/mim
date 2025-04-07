package tui

import (
    "math"

    "github.com/NimbleMarkets/ntcharts/heatmap"
    "github.com/jonsampson/mim/internal/domain"
)

type CPUHeatmap struct {
    hm              *heatmap.Model
    squareDimension int
    metrics         domain.CPUMemoryMetrics
}

func NewCPUHeatmap() *CPUHeatmap {
    return &CPUHeatmap{}
}

func (c *CPUHeatmap) Update(msg interface{}) {
    switch msg := msg.(type) {
    case domain.CPUMemoryMetrics:
        c.metrics = msg
        c.updateHeatmap()
    }
}

func (c *CPUHeatmap) updateHeatmap() {
    c.squareDimension = int(math.Ceil(math.Sqrt(float64(len(c.metrics.CPUUsagePerCore)))))
    heatMap := heatmap.New(c.squareDimension+1,
        c.squareDimension+1,
        heatmap.WithValueRange(0, 100),
    )
    
    matrix := make([][]float64, c.squareDimension)
    for i := range matrix {
        matrix[i] = make([]float64, c.squareDimension)
    }

    core := 0
    for i := 0; i < c.squareDimension; i++ {
        for j := 0; j < c.squareDimension; j++ {
            if core < len(c.metrics.CPUUsagePerCore) {
                // Rotate 90 degrees clockwise
                matrix[j][c.squareDimension-1-i] = c.metrics.CPUUsagePerCore[core]
                core++
            } else {
                matrix[j][c.squareDimension-1-i] = 0
            }
        }
    }

    heatMap.PushAllMatrixRow(matrix)
    c.hm = &heatMap
}

func (c *CPUHeatmap) View() string {
    if c.hm == nil {
        return ""
    }
    c.hm.Draw()
    return c.hm.View()
}