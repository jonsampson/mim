package tui

import (
	"fmt"
	"log"

	"github.com/NimbleMarkets/ntcharts/sparkline"
)

type CoreSparkline struct {
	sparkline *sparkline.Model
	coreID    int
	lastValue float64
	width    int
	height   int
}

func NewCoreSparkline(coreID int) *CoreSparkline {
	sparkline := sparkline.New(
		20,
		1,
		sparkline.WithMaxValue(100),
	)
	coreSparkline := &CoreSparkline{
		coreID:    coreID,
		lastValue: 0,
		sparkline: &sparkline,
	}
	return coreSparkline
}

func (c *CoreSparkline) Update(value float64) {
	c.lastValue = value
	c.sparkline.Push(value)
}

func (c *CoreSparkline) View() string {
	if c.sparkline == nil {
		return ""
	}
	return fmt.Sprintf("Core %2d: %s [%5.1f%%]", c.coreID, c.sparkline.View(), c.lastValue)
}

func (c *CoreSparkline) Resize(width, height int) {
	// Adjust the sparkline size based on available space
	// This is a simplified example; you might want to use a more sophisticated calculation
	c.width = width
	c.height = height
	log.Printf("Resizing sparkline for core %d to width %d", c.coreID, width-23)
	c.sparkline.Resize(width-23, 1)
}
