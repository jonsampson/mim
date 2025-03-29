package main

import (
    "fmt"
    "os"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/jonsampson/mim/internal/infra"
    "github.com/jonsampson/mim/internal/tui"
)

func main() {
    factory := infra.CollectorFactory{}
    collectors := factory.CreateCollectors()

    model, err := tui.InitialModel(collectors...)
    if err != nil {
        fmt.Printf("Error initializing model: %v\n", err)
        os.Exit(1)
    }

    p := tea.NewProgram(model)
    if _, err := p.Run(); err != nil {
        fmt.Printf("Alas, there's been an error: %v", err)
        os.Exit(1)
    }
}
