package tui

import (
    "fmt"

    "github.com/NimbleMarkets/ntcharts/sparkline"
)

type CoreSparkline struct {
    sparkline *sparkline.Model
    coreID    int
    lastValue float64
    width     int
    height    int
}

func NewCoreSparkline(coreID int) *CoreSparkline {
    sparkline := sparkline.New(
        1,
        1,
        sparkline.WithMaxValue(100),
    )
    return &CoreSparkline{
        coreID:    coreID,
        lastValue: 0,
        sparkline: &sparkline,
        width:     20, // Default width
        height:    1,  // Default height
    }
}

func (c *CoreSparkline) Update(msg interface{}) {
    switch msg := msg.(type) {
    case float64:
        c.updateValue(msg)
    }
}

func (c *CoreSparkline) updateValue(value float64) {
    c.lastValue = value
    if c.sparkline != nil {
        c.sparkline.Push(value)
    }
}

func (c *CoreSparkline) View() string {
    if c.sparkline == nil {
        return ""
    }
    c.sparkline.Draw() // Ensure the sparkline is drawn before viewing
    return fmt.Sprintf("Core %2d: %s [%5.1f%%]", c.coreID, c.sparkline.View(), c.lastValue)
}

func (c *CoreSparkline) Resize(width, height int) {
    c.width = width
    c.height = height
    if c.sparkline != nil {
        // Adjust the sparkline size based on available space
        // This is a simplified example; you might want to use a more sophisticated calculation
        c.sparkline.Resize(width-17, 1)
    }
}