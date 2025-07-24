package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jonsampson/mim/internal/domain"
	"github.com/shirou/gopsutil/v4/process"
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
	// Pre-allocated buffers for string formatting
	rowBuffer       []table.Row
	strBuilder      strings.Builder
}

const (
	symbolWidth    = 6
	pidWidth       = 12
	userWidth      = 12
	metricWidth    = 12
	minCommandWidth = 20 // Minimum viable command column width
)

func NewProcessMonitor(width int) *ProcessMonitor {
	pm := &ProcessMonitor{
		width:           width,
		symbolAllocator: NewSymbolAllocator([]rune{'▣', '▤', '▥', '▦', '▧', '▨', '▩', '▪', '▫', '▬', '◆', '◇', '○', '●', '◉', '◍', '◎', '◌', '◔', '◕'}),
		borderStyle:     lipgloss.NewStyle().Padding(0).Margin(0),
		rowBuffer:       make([]table.Row, 0, 5), // Pre-allocate for 5 rows
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
	commandWidth := (width - symbolWidth - pidWidth - userWidth - metricWidth)
	columns := []table.Column{
		{Title: "   Key", Width: symbolWidth},
		{Title: "         PID", Width: pidWidth},
		{Title: "        User", Width: userWidth},
		{Title: "         %", Width: metricWidth},
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
	hexBuf := make([]byte, 7) // "#RRGGBB"
	hexBuf[0] = '#'
	
	for i := range colors {
		hue := float64(i) / float64(count) * 360.0
		r, g, b := hslToRGB(hue, 1.0, 0.5)
		
		// Manual hex formatting to avoid fmt.Sprintf
		hexChars := "0123456789abcdef"
		hexBuf[1] = hexChars[r>>4]
		hexBuf[2] = hexChars[r&0xf]
		hexBuf[3] = hexChars[g>>4]
		hexBuf[4] = hexChars[g&0xf]
		hexBuf[5] = hexChars[b>>4]
		hexBuf[6] = hexChars[b&0xf]
		
		colors[i] = lipgloss.NewStyle().Foreground(lipgloss.Color(string(hexBuf)))
	}
	return colors
}

func (pm *ProcessMonitor) Init() tea.Cmd {
	return nil
}

func (pm *ProcessMonitor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	pm.cpuTable, cmd = pm.cpuTable.Update(msg)
	cmds = append(cmds, cmd)
	pm.memTable, cmd = pm.memTable.Update(msg)
	cmds = append(cmds, cmd)
	pm.gpuTable, cmd = pm.gpuTable.Update(msg)
	cmds = append(cmds, cmd)
	pm.gpuMemTable, cmd = pm.gpuMemTable.Update(msg)
	cmds = append(cmds, cmd)
	return pm, tea.Batch(cmds...)
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

	// Calculate minimum width needed for 2x2 layout
	minTableWidth := symbolWidth + pidWidth + userWidth + metricWidth + minCommandWidth
	paddingWidth := 6 // Account for borders and padding between tables
	min2x2Width := 2*minTableWidth + paddingWidth

	// Use responsive layout based on available width
	var view string
	if pm.width >= min2x2Width {
		// Wide screen: use 2x2 grid
		topRow := lipgloss.JoinHorizontal(lipgloss.Top, padding(cpuView), memView)
		bottomRow := lipgloss.JoinHorizontal(lipgloss.Top, padding(gpuView), gpuMemView)
		view = lipgloss.JoinVertical(lipgloss.Left, topRow, bottomRow)
	} else {
		// Narrow screen: stack vertically
		view = lipgloss.JoinVertical(lipgloss.Left, cpuView, memView, gpuView, gpuMemView)
	}

	return lipgloss.NewStyle().Border(lipgloss.HiddenBorder()).Render(view)
}

func (pm *ProcessMonitor) UpdateProcesses(cpuProcesses []domain.CPUProcessInfo, gpuProcesses []domain.GPUProcessInfo) {
	pm.cpuProcesses = cpuProcesses
	pm.gpuProcesses = gpuProcesses

	// Create a map of PID to Command from CPU processes for GPU command lookup
	pidToCommandForGPU := make(map[uint32]string)
	for _, p := range pm.cpuProcesses {
		pidToCommandForGPU[p.Pid] = p.Command
	}

	// Update CPU table - sort by CPU%
	sort.Slice(pm.cpuProcesses, func(i, j int) bool {
		return pm.cpuProcesses[i].CPUPercent > pm.cpuProcesses[j].CPUPercent
	})
	pm.cpuTable.SetRows(pm.getRows(pm.cpuProcesses, func(p domain.CPUProcessInfo) float64 { return p.CPUPercent }))

	// Update MEM table - sort by Memory%
	sort.Slice(pm.cpuProcesses, func(i, j int) bool {
		return pm.cpuProcesses[i].MemoryPercent > pm.cpuProcesses[j].MemoryPercent
	})
	pm.memTable.SetRows(pm.getRows(pm.cpuProcesses, func(p domain.CPUProcessInfo) float64 { return p.MemoryPercent }))

	// Update GPU table - sort by GPU%
	sort.Slice(pm.gpuProcesses, func(i, j int) bool {
		return pm.gpuProcesses[i].SmUtil > pm.gpuProcesses[j].SmUtil
	})
	pm.gpuTable.SetRows(pm.getGPURows(pm.gpuProcesses, pidToCommandForGPU, func(p domain.GPUProcessInfo) float64 { return float64(p.SmUtil) }))

	// Update GPU MEM table - sort by GPU Memory
	sort.Slice(pm.gpuProcesses, func(i, j int) bool {
		return pm.gpuProcesses[i].UsedGpuMemory > pm.gpuProcesses[j].UsedGpuMemory
	})
	pm.gpuMemTable.SetRows(pm.getGPURows(pm.gpuProcesses, pidToCommandForGPU, func(p domain.GPUProcessInfo) float64 { return p.UsedGpuMemory }))
}

func (pm *ProcessMonitor) getRows(processes []domain.CPUProcessInfo, getValue func(domain.CPUProcessInfo) float64) []table.Row {
	// Reuse the pre-allocated buffer
	pm.rowBuffer = pm.rowBuffer[:0]
	
	for i := range min(5, len(processes)) {
		p := processes[i]
		sym, _ := pm.symbolAllocator.AccessPID(int(p.Pid))
		
		// Get username if not already populated (expensive operation)
		user := p.User
		if user == "" {
			if proc, err := process.NewProcess(int32(p.Pid)); err == nil {
				if username, err := proc.Username(); err == nil {
					user = username
				} else {
					user = "?"
				}
			} else {
				user = "?"
			}
		}
		
		pm.rowBuffer = append(pm.rowBuffer, table.Row{
			pm.formatSymbol(sym),
			pm.formatPID(p.Pid),
			pm.formatUser(user),
			pm.formatMetric(getValue(p)),
			p.Command,
		})
	}
	return pm.rowBuffer
}

func (pm *ProcessMonitor) getGPURows(processes []domain.GPUProcessInfo, pidToCommand map[uint32]string, getValue func(domain.GPUProcessInfo) float64) []table.Row {
	// Reuse the pre-allocated buffer
	pm.rowBuffer = pm.rowBuffer[:0]
	
	for i := range min(5, len(processes)) {
		p := processes[i]
		sym, _ := pm.symbolAllocator.AccessPID(int(p.Pid))
		command := pidToCommand[p.Pid]
		
		pm.rowBuffer = append(pm.rowBuffer, table.Row{
			pm.formatSymbol(sym),
			pm.formatPID(p.Pid),
			pm.formatUser(p.User),
			pm.formatMetric(getValue(p)),
			command,
		})
	}
	return pm.rowBuffer
}


// Formatting helper methods to reduce allocations
func (pm *ProcessMonitor) formatSymbol(sym rune) string {
	pm.strBuilder.Reset()
	pm.strBuilder.WriteString("     ")
	pm.strBuilder.WriteRune(sym)
	return pm.strBuilder.String()
}

func (pm *ProcessMonitor) formatPID(pid uint32) string {
	return fmt.Sprintf("%12d", pid)
}

func (pm *ProcessMonitor) formatUser(user string) string {
	if len(user) > 12 {
		user = user[:12]
	}
	return fmt.Sprintf("%12s", user)
}

func (pm *ProcessMonitor) formatMetric(value float64) string {
	return fmt.Sprintf("%10.1f", value)
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
	
	// Calculate minimum width needed for 2x2 layout
	minTableWidth := symbolWidth + pidWidth + userWidth + metricWidth + minCommandWidth
	paddingWidth := 6 // Account for borders and padding between tables
	min2x2Width := 2*minTableWidth + paddingWidth

	var tableWidth int
	if width >= min2x2Width {
		// Wide screen: use half width for 2x2 grid
		tableWidth = width/2 - 4
	} else {
		// Narrow screen: use full width for vertical stack
		tableWidth = width - 4
	}

	updateTable := func(t *table.Model) {
		*t = pm.createTableWithWidth(tableWidth)
	}

	updateTable(&pm.cpuTable)
	updateTable(&pm.memTable)
	updateTable(&pm.gpuTable)
	updateTable(&pm.gpuMemTable)
}
