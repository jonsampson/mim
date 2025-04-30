package tui

import (
    "fmt"
    "sort"

    "github.com/charmbracelet/bubbles/table"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "github.com/jonsampson/mim/internal/domain"
)

type ProcessMonitor struct {
    cpuProcesses []domain.CPUProcessInfo
    gpuProcesses []domain.GPUProcessInfo
    cpuTable     table.Model
    gpuTable     table.Model
}

func NewProcessMonitor() *ProcessMonitor {
    cpuColumns := []table.Column{
        {Title: "PID", Width: 10},
        {Title: "CPU%", Width: 10},
        {Title: "MEM%", Width: 10},
        {Title: "Command", Width: 30},
    }

    gpuColumns := []table.Column{
        {Title: "PID", Width: 10},
        {Title: "GPU%", Width: 10},
        {Title: "GPU MEM", Width: 10},
        {Title: "Command", Width: 30},
    }

    cpuTable := table.New(
        table.WithColumns(cpuColumns),
        table.WithFocused(true),
        table.WithHeight(6),
    )

    gpuTable := table.New(
        table.WithColumns(gpuColumns),
        table.WithFocused(true),
        table.WithHeight(6),
    )

    s := table.DefaultStyles()
    s.Header = s.Header.
        BorderStyle(lipgloss.HiddenBorder()).
        BorderBottom(true)
    s.Selected = s.Selected.
        Foreground(lipgloss.Color("229")).
        Background(lipgloss.Color("57")).
        Bold(false)
    cpuTable.SetStyles(s)
    gpuTable.SetStyles(s)

    return &ProcessMonitor{
        cpuTable: cpuTable,
        gpuTable: gpuTable,
    }
}

func (pm *ProcessMonitor) Init() tea.Cmd {
    return nil
}

func (pm *ProcessMonitor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    pm.cpuTable, cmd = pm.cpuTable.Update(msg)
    pm.gpuTable, _ = pm.gpuTable.Update(msg)
    return pm, cmd
}

func (pm *ProcessMonitor) View() string {
    return lipgloss.JoinHorizontal(
        lipgloss.Top,
        lipgloss.NewStyle().Margin(0, 1).Render("CPU Processes\n"+pm.cpuTable.View()),
        lipgloss.NewStyle().Margin(0, 1).Render("GPU Processes\n"+pm.gpuTable.View()),
    )
}

func (pm *ProcessMonitor) UpdateProcesses(cpuProcesses []domain.CPUProcessInfo, gpuProcesses []domain.GPUProcessInfo) {
    pm.cpuProcesses = cpuProcesses
    pm.gpuProcesses = gpuProcesses

    // Sort CPU processes
    sort.Slice(pm.cpuProcesses, func(i, j int) bool {
        return pm.cpuProcesses[i].CPUPercent > pm.cpuProcesses[j].CPUPercent
    })

    // Update CPU table
    cpuRows := []table.Row{}
    for i := 0; i < min(5, len(pm.cpuProcesses)); i++ {
        p := pm.cpuProcesses[i]
        cpuRows = append(cpuRows, table.Row{
            fmt.Sprintf("%d", p.Pid),
            fmt.Sprintf("%.1f", p.CPUPercent),
            fmt.Sprintf("%.1f", p.MemoryPercent),
            p.Command,
        })
    }
    pm.cpuTable.SetRows(cpuRows)

    // Create a map of PID to Command from CPU processes
    pidToCommand := make(map[uint32]string)
    for _, p := range pm.cpuProcesses {
        pidToCommand[p.Pid] = p.Command
    }

    // Sort GPU processes
    sort.Slice(pm.gpuProcesses, func(i, j int) bool {
        return pm.gpuProcesses[i].SmUtil > pm.gpuProcesses[j].SmUtil
    })

    // Update GPU table
    gpuRows := []table.Row{}
    for i := 0; i < min(5, len(pm.gpuProcesses)); i++ {
        p := pm.gpuProcesses[i]
        command := pidToCommand[p.Pid] // Get command from CPU processes
        gpuRows = append(gpuRows, table.Row{
            fmt.Sprintf("%d", p.Pid),
            fmt.Sprintf("%.1f", float64(p.SmUtil)),
            fmt.Sprintf("%.1f", float64(p.UsedGpuMemory)/(1024*1024)), // Convert to MB
            command,
        })
    }
    pm.gpuTable.SetRows(gpuRows)
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
