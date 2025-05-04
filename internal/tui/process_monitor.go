package tui

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jonsampson/mim/internal/domain"
)

type ProcessMonitor struct {
	cpuProcesses    []domain.CPUProcessInfo
	gpuProcesses    []domain.GPUProcessInfo
	cpuTable        table.Model
	memTable        table.Model
	gpuTable        table.Model
	gpuMemTable     table.Model
	symbolAllocator *SymbolAllocator
	symbolColors    []lipgloss.Style
	width           int
	borderStyle     lipgloss.Style
}

const (
	symbolWidth = 6
	pidWidth    = 12
	metricWidth = 12
)

func NewProcessMonitor(width int) *ProcessMonitor {
	pm := &ProcessMonitor{
		width:           width,
		symbolAllocator: NewSymbolAllocator([]rune{'▣', '▤', '▥', '▦', '▧', '▨', '▩', '▪', '▫', '▬', '◆', '◇', '○', '●', '◉', '◍', '◎', '◌', '◔', '◕'}),
		borderStyle:     lipgloss.NewStyle().Border(lipgloss.HiddenBorder()),
	}

	pm.symbolColors = createSymbolColors(len(pm.symbolAllocator.symbols))
	pm.cpuTable = pm.createTable()
	pm.memTable = pm.createTable()
	pm.gpuTable = pm.createTable()
	pm.gpuMemTable = pm.createTable()

	return pm
}

func (pm *ProcessMonitor) createTable() table.Model {
	return pm.createTableWithWidth(pm.width/2 - 4)
}

func (pm *ProcessMonitor) createTableWithWidth(width int) table.Model {
	commandWidth := (width - symbolWidth - pidWidth - metricWidth)
	columns := []table.Column{
		{Title: fmt.Sprintf("%6s", "Key"), Width: symbolWidth},
		{Title: fmt.Sprintf("%12s", "PID"), Width: pidWidth},
		{Title: fmt.Sprintf("%10s", "%"), Width: metricWidth},
		{Title: "Command", Width: commandWidth},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(false),
		table.WithHeight(6), // 1 header + 5 rows
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false).
		Padding(0).
		Margin(0)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("default")).
		Background(lipgloss.Color("default")).
		Bold(false).
		Padding(0).Margin(0)
	s.Cell = s.Cell.
		Padding(0).Margin(0)

	t.SetStyles(s)
	return t
}

func createSymbolColors(count int) []lipgloss.Style {
	colors := make([]lipgloss.Style, count)
	for i := range colors {
		hue := float64(i) / float64(count) * 360.0
		r, g, b := hslToRGB(hue, 1.0, 0.5)
		colors[i] = lipgloss.NewStyle().Foreground(lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", r, g, b)))
	}
	return colors
}

func (pm *ProcessMonitor) Init() tea.Cmd {
	return nil
}

func (pm *ProcessMonitor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	pm.cpuTable, cmd = pm.cpuTable.Update(msg)
	pm.memTable, _ = pm.memTable.Update(msg)
	pm.gpuTable, _ = pm.gpuTable.Update(msg)
	pm.gpuMemTable, _ = pm.gpuMemTable.Update(msg)
	return pm, cmd
}

func (pm *ProcessMonitor) View() string {
	padding := lipgloss.NewStyle().PaddingRight(2).Render

	cpuView := pm.borderStyle.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		"CPU %",
		pm.cpuTable.View(),
	))

	memView := pm.borderStyle.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		"MEM %",
		pm.memTable.View(),
	))

	gpuView := pm.borderStyle.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		"GPU %",
		pm.gpuTable.View(),
	))

	gpuMemView := pm.borderStyle.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		"GPU MEM",
		pm.gpuMemTable.View(),
	))

	topRow := lipgloss.JoinHorizontal(lipgloss.Top, padding(cpuView), memView)
	bottomRow := lipgloss.JoinHorizontal(lipgloss.Top, padding(gpuView), gpuMemView)

	view := lipgloss.JoinVertical(lipgloss.Left, topRow, bottomRow)

	return pm.borderStyle.Render(view)
}

func (pm *ProcessMonitor) UpdateProcesses(cpuProcesses []domain.CPUProcessInfo, gpuProcesses []domain.GPUProcessInfo) {
	pm.cpuProcesses = cpuProcesses
	pm.gpuProcesses = gpuProcesses

	// Create a map of PID to Command from CPU processes
	pidToCommand := make(map[uint32]string)
	for _, p := range pm.cpuProcesses {
		pidToCommand[p.Pid] = p.Command
	}

	// Update CPU table
	sort.Slice(pm.cpuProcesses, func(i, j int) bool {
		return pm.cpuProcesses[i].CPUPercent > pm.cpuProcesses[j].CPUPercent
	})
	pm.cpuTable.SetRows(pm.getRows(pm.cpuProcesses, func(p domain.CPUProcessInfo) float64 { return p.CPUPercent }))

	// Update MEM table
	sort.Slice(pm.cpuProcesses, func(i, j int) bool {
		return pm.cpuProcesses[i].MemoryPercent > pm.cpuProcesses[j].MemoryPercent
	})
	pm.memTable.SetRows(pm.getRows(pm.cpuProcesses, func(p domain.CPUProcessInfo) float64 { return p.MemoryPercent }))

	// Update GPU table
	sort.Slice(pm.gpuProcesses, func(i, j int) bool {
		return pm.gpuProcesses[i].SmUtil > pm.gpuProcesses[j].SmUtil
	})
	pm.gpuTable.SetRows(pm.getGPURows(pm.gpuProcesses, pidToCommand, func(p domain.GPUProcessInfo) float64 { return float64(p.SmUtil) }))

	// Update GPU MEM table
	sort.Slice(pm.gpuProcesses, func(i, j int) bool {
		return pm.gpuProcesses[i].UsedGpuMemory > pm.gpuProcesses[j].UsedGpuMemory
	})
	pm.gpuMemTable.SetRows(pm.getGPURows(pm.gpuProcesses, pidToCommand, func(p domain.GPUProcessInfo) float64 { return float64(p.UsedGpuMemory) / (1024 * 1024) }))
}

func (pm *ProcessMonitor) getRows(processes []domain.CPUProcessInfo, getValue func(domain.CPUProcessInfo) float64) []table.Row {
	rows := []table.Row{}
	for i := 0; i < min(5, len(processes)); i++ {
		p := processes[i]
		sym, _ := pm.symbolAllocator.AccessPID(int(p.Pid))
		// debugInfo := fmt.Sprintf("S:%c C:%d", sym, colorIndex)
		rows = append(rows, table.Row{
			fmt.Sprintf("%6c", sym),
			fmt.Sprintf("%12s", strconv.FormatUint(uint64(p.Pid), 10)),
			fmt.Sprintf("%10.1f", getValue(p)),
			p.Command,
		})
	}
	return rows
}

func (pm *ProcessMonitor) getGPURows(processes []domain.GPUProcessInfo, pidToCommand map[uint32]string, getValue func(domain.GPUProcessInfo) float64) []table.Row {
	rows := []table.Row{}
	for i := range min(5, len(processes)) {
		p := processes[i]
		sym, _ := pm.symbolAllocator.AccessPID(int(p.Pid))
		// debugInfo := fmt.Sprintf("S:%c C:%d", sym, colorIndex)
		command := pidToCommand[p.Pid]
		rows = append(rows, table.Row{
			fmt.Sprintf("%6c", sym),
			fmt.Sprintf("%12s", strconv.FormatUint(uint64(p.Pid), 10)),
			fmt.Sprintf("%10.1f", getValue(p)),
			command,
		})
	}
	return rows
}

// Helper function to convert HSL to RGB
func hslToRGB(h, s, l float64) (r, g, b uint8) {
	if s == 0 {
		r = uint8(l * 255)
		g = uint8(l * 255)
		b = uint8(l * 255)
		return
	}

	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q

	r = uint8(hueToRGB(p, q, h+1.0/3.0) * 255)
	g = uint8(hueToRGB(p, q, h) * 255)
	b = uint8(hueToRGB(p, q, h-1.0/3.0) * 255)
	return
}

func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6*t
	}
	if t < 1.0/2.0 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6
	}
	return p
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (pm *ProcessMonitor) Resize(width int) {
	pm.width = width
	tableWidth := width/2 - 13 // Adjusted to account for borders

	updateTable := func(t *table.Model) {
		*t = pm.createTableWithWidth(tableWidth)
	}

	updateTable(&pm.cpuTable)
	updateTable(&pm.memTable)
	updateTable(&pm.gpuTable)
	updateTable(&pm.gpuMemTable)
}
