package tui

import (
    "strings"

    "github.com/jonsampson/mim/internal/domain"
)

// CPUCombinedView combines the CPU heatmap and busiest cores view side by side
type CPUCombinedView struct {
    heatmap      *CPUHeatmap
    busiestCores *BusiestCores
    width        int
    height       int
}

// NewCPUCombinedView creates a new combined view of CPU heatmap and busiest cores
func NewCPUCombinedView() *CPUCombinedView {
    return &CPUCombinedView{
        heatmap:      NewCPUHeatmap(),
        busiestCores: NewBusiestCores(),
    }
}

// Update updates both the heatmap and busiest cores with new metrics
func (c *CPUCombinedView) Update(metrics domain.CPUMemoryMetrics) {
    c.heatmap.Update(metrics)
    c.busiestCores.Update(metrics)
}

// View renders the heatmap and busiest cores side by side
func (c *CPUCombinedView) View() string {
    heatmapView := strings.Split(c.heatmap.View(), "\n")
    busiestView := strings.Split(c.busiestCores.View(), "\n")
    
    // Determine the maximum number of lines
    maxLines := len(heatmapView)
    if len(busiestView) > maxLines {
        maxLines = len(busiestView)
    }
    
    // Pad the shorter view with empty lines
    for len(heatmapView) < maxLines {
        heatmapView = append(heatmapView, "")
    }
    for len(busiestView) < maxLines {
        busiestView = append(busiestView, "")
    }
    
    // Combine the views side by side
    var combined []string
    for i := 0; i < maxLines; i++ {
        combined = append(combined, heatmapView[i]+" "+busiestView[i])
    }
    
    return strings.Join(combined, "\n")
}

// Resize adjusts the size of both components
func (c *CPUCombinedView) Resize(width, height int) {
    c.width = width
    c.height = height
    c.busiestCores.Resize(width, height)
}