# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Mim (Monitoring is Mim) is a terminal-based system resource monitoring application that displays CPU, GPU, memory usage, and process information in a single TUI screen.

## Architecture

The project follows Clean Architecture with the Elm Architecture pattern using Bubble Tea:

- **Domain Layer** (`internal/domain/`): Core data structures for metrics
- **Infrastructure Layer** (`internal/infra/`): System metric collectors using gopsutil and go-nvml
- **TUI Layer** (`internal/tui/`): UI components following Model-Update-View pattern
- **Entry Point** (`cmd/mim/`): Main application initialization

Key architectural points:
- Unidirectional data flow with Messages -> Update -> Model -> View
- Collectors run in separate goroutines sending metrics via channels
- Factory pattern for creating appropriate collectors (CPU/Memory, GPU)
- Dependency injection for testability with mock implementations

## Development Commands

```bash
# Build the application
go build ./cmd/mim/...

# Run tests
go test ./...

# Install dependencies
go mod download

# Tidy dependencies
go mod tidy

# Run the application
go run cmd/mim/main.go

# Profile the application
./mim -cpuprofile=cpu.prof     # CPU profile to file
./mim -webpprof                 # Web-based profiling on :6060
./profile.sh                    # Interactive profiling helper
```

## Key Implementation Details

1. **Collector Pattern**: All collectors implement `MetricsCollector` interface with `Start()` and `Stop()` methods
2. **Process Monitoring**: Collects per-process CPU%, memory%, and GPU usage with user information
3. **GPU Support**: Currently NVIDIA-only via go-nvml, AMD support planned
4. **Testing**: Uses testify for assertions, mock collectors in `mocks.go` files
5. **Debug Logging**: Writes to `debug.log` file for troubleshooting

## Important Files

- `internal/tui/model.go`: Core application state and update logic
- `internal/infra/collector_factory.go`: Creates appropriate collectors based on system capabilities
- `internal/domain/metrics.go`: Data structures for all metric types

## Dependencies

- Go 1.23.3 (minimum 1.18)
- charmbracelet/bubbletea: TUI framework
- shirou/gopsutil/v3: CPU/memory metrics
- NVIDIA/go-nvml: GPU metrics
- NimbleMarkets/ntcharts: Graphing components

## Bubbletea Component Architecture

### Component Hierarchy
- Model (root) contains all sub-components
- CPUCombinedView contains CPUHeatmap and BusiestCores
- ProcessMonitor contains 4 table.Model instances and SymbolAllocator
- Maximum nesting depth: 3 levels

### State Flow
- Messages: tea.KeyMsg, tea.WindowSizeMsg, domain.CPUMemoryMetrics, domain.GPUMetrics
- Circular pattern: Metrics arrive → Update state → listenForMetrics → Repeat
- Terminal states: "q"/"ctrl+c" (os.Exit) or "end" (tea.Quit)
- All components re-render on every update (no selective rendering)

### Performance Considerations
1. **Redundant rendering**: All components re-render even if data unchanged
2. **String operations**: Heavy concatenation in renderContent()
3. **Blocking collection**: CPU sampling blocks for 1 second
4. **No view caching**: Components lack change detection

### Key Patterns
- No message bubbling (all messages handled at root)
- Components are pure display elements without own Update cycles
- Simple command pattern: Initial batch + recursive metric listening
- Shared state via SymbolAllocator for PID→symbol mappings

### View() Performance Characteristics

#### String Allocations (per frame)
- ProcessMonitor: ~20 fmt.Sprintf calls (optimized from ~100)
- BusiestCores: ~16-32 fmt.Sprintf calls
- Total: ~40-60 string allocations per render (optimized from ~150-200)
- Heavy lipgloss.Join operations: ~10 per frame

#### Remaining Performance Issues
1. **No Component Caching**: All components re-render on any update (CPU/GPU/resize)
2. **O(n²) Complexity**: BusiestCores uses nested loops for rendering
3. **String Concatenation**: Using + operator in loops instead of strings.Builder

#### Hotspots by Component Size
1. ProcessMonitor (~150 lines): 4 tables, optimized string formatting
2. BusiestCores (~80 lines): O(n²) loops for rendering
3. CPUCombinedView (~40 lines): Multiple Split/Join operations
4. CPUHeatmap (~60 lines): Recreates entire heatmap model on update

#### Remaining Optimization Opportunities
- Implement dirty-flag pattern for selective rendering
- Use strings.Builder for loop concatenations  
- Reuse heatmap model instead of recreating

#### Completed Optimizations

**BusiestCores Style Caching**
- Cached lipgloss styles at initialization instead of creating per frame
- Reduced from 48 style objects per frame to 4 per initialization

**ProcessMonitor String Optimization**  
- Pre-allocated row buffer to reuse across updates
- Removed fmt.Sprintf from column headers  
- Optimized color generation to avoid fmt.Sprintf
- Added string builder for efficient concatenation
- Reduced fmt.Sprintf calls by ~80% (100 → 20 per frame)
- Implemented index-based sorting to eliminate redundant sorts (4x → 1x per data type)

**CPUMemoryCollector Performance Overhaul**
- **Eliminated blocking `CPUPercent()` calls**: Replaced with delta calculation using `Times()`
- **Single `/proc` read per process**: Calculate CPU% from stored deltas between collection cycles
- **Reduced process collection CPU usage**: From 73% to 35% of total CPU time
- **Limited expensive operations**: Memory% for top 25, username for top 10 processes only
- **Non-blocking approach**: No sleep calls, uses natural 1-second collection interval

**Profiling Infrastructure**
- Added pprof support with `-cpuprofile` and `-webpprof` flags
- Created profiling documentation and helper scripts
- Proper signal handling to ensure profile data is written on exit