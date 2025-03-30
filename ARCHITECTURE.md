# Mim Architecture & Development Guide

## Overview
Mim is a terminal-based user interface (TUI) application designed to monitor system resources, including CPU usage (aggregate and per core), GPU usage, and memory consumption (both system and GPU). The goal is to provide a comprehensive, single-screen view of system performance, reducing the need for multiple tools like `top`, `bashtop`, and `nvtop`.

## Technical Architecture

### Core Technologies
- **Language:** Go
- **TUI Framework:** [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **Graphing Library:** [ntcharts](https://github.com/charmbracelet/bubbletea/tree/master/examples/charts)
- **Metrics Collection:**
  - **CPU & Memory:** [`gopsutil`](https://github.com/shirou/gopsutil)
  - **NVIDIA GPU:** [`nvml`](https://github.com/NVIDIA/go-nvml)
  - **AMD GPU:** [`go_amd_smi`](https://github.com/amd/go_amd_smi)

### Data Flow
Mim follows the **Elm Architecture**, as enforced by Bubble Tea:
1. **Model:** Maintains application state.
2. **Messages (Msg):** Events that trigger updates.
3. **Update Function:** Processes messages and updates the model.
4. **View Function:** Renders the UI based on the model.

To efficiently update the model, Mim will use **Go channels** to pass system metrics updates from data collection routines to the Bubble Tea event loop.

## Project Structure

```
mim/
├── cmd/          # Main entry point
│   ├── mim/      # CLI setup, Bubble Tea initialization
│   └── ...
├── internal/     # Core application logic
│   ├── domain/   # Entities and interfaces (e.g., Process, Metrics, Repository)
│   ├── service/  # Business logic (e.g., MetricsCollector, ProcessAnalyzer)
│   ├── tui/      # UI components (Bubble Tea model, views, messages)
│   ├── infra/    # External dependencies (e.g., gopsutil, nvml, go_amd_smi wrappers)
│   └── mocks/    # Mocks for testing
├── pkg/          # Potential reusable components
├── scripts/      # Dev and build scripts
└── tests/        # Integration tests
```

## AI-Accelerated Development
To enhance productivity and maintainability, Mim will integrate AI-assisted development practices:
- **Code Suggestions & Refactoring:** Using AI tools to streamline coding and detect potential optimizations.
- **Automated Testing:** AI-driven test generation to improve coverage.
- **Documentation & Code Reviews:** Leveraging AI to suggest improvements in readability and maintainability.
- **Mock Generation:** Automating the creation of `mocks.go` files for unit tests.

## Testing Strategy
Mim will follow **Clean Architecture** principles to ensure modularity and testability:
- **Inversion of Control:** Dependencies will be injected via interfaces to decouple components.
- **Private Interfaces:** Defining behavior in a way that allows easy substitution for testing.
- **Mocking Approach:**
  - Mocks will be placed in a `mocks.go` file within each relevant package.
  - Unit tests will use mocks to isolate components.

## Future Considerations

### AMD GPU Metrics Collection

While the current implementation focuses on CPU, Memory, and NVIDIA GPU metrics, we've identified the need to support AMD GPU metrics in the future. The integration of the `goamdsmi` library for AMD GPU metrics collection has presented some challenges:

1. Compatibility issues with the current project structure.
2. Potential complexities in cross-platform support.
3. Need for more extensive testing on systems with AMD GPUs.

As a result, we've decided to postpone the implementation of AMD GPU metrics collection. This feature will be revisited in future iterations of the project. When implementing, consider the following:

- Ensure proper error handling for systems without AMD GPUs.
- Implement a fallback mechanism for generic GPU information if specific AMD metrics are unavailable.
- Consider creating a more generic GPU interface that can be implemented by both NVIDIA and AMD collectors.

TODO: Revisit AMD GPU metrics collection implementation using `goamdsmi` or alternative libraries.

## Next Steps

1) Define the initial static UI layout (grid structure for CPU, GPU, and memory monitoring).

Here's the proposed layout for the initial static UI:

```
+----------------------------------+----------------------------------+
|             CPU Usage            |           Memory Usage           |
|                                  |                                  |
| [Aggregate CPU Usage Bar/Graph]  | [System Memory Usage Bar/Graph]  |
|                                  |                                  |
| [CPU Core Heatmap]               | [Memory Usage Details]           |
| (Showing all cores as color-     | - Total:                         |
|  coded squares, with color       | - Used:                          |
|  intensity representing usage)   | - Free:                          |
|                                  | - Cached:                        |
| Top 5 Busiest Cores:             |                                  |
| Core 7:  [===========]  90%      | [Swap Usage Bar/Graph]           |
| Core 12: [==========]   85%      | - Total:                         |
| Core 3:  [=========]    80%      | - Used:                          |
| Core 18: [========]     75%      | - Free:                          |
| Core 1:  [=======]      70%      |                                  |
+----------------------------------+----------------------------------+
|             GPU Usage            |         Process Monitor          |
|                                  |                                  |
| [GPU Usage Bar/Graph]            | [Top Processes Table]            |
|                                  |                                  |
| [GPU Memory Usage Bar/Graph]     | PID  CPU%  MEM%  Command         |
|                                  | 1234  25.0  10.2  process1       |
| GPU Details:                     | 5678  15.5   8.7  process2       |
| - Model:                         | 9012  10.2   5.3  process3       |
| - Temperature:                   | 3456   8.7   4.1  process4       |
| - Fan Speed:                     | 7890   7.3   3.8  process5       |
| - Power Usage:                   |                                  |
+----------------------------------+----------------------------------+
```

Implementation Notes:

a) CPU Usage Section:
   - Implement a custom Bubble Tea component for the CPU Core Heatmap.
   - Use a grid of colored blocks, where each block represents a core and its color intensity represents usage.
   - Implement a sorting mechanism to identify and display the top 5 busiest cores.

b) Memory Usage Section:
   - Use Bubble Tea's built-in components or create custom ones for the memory usage bar graphs.
   - Implement a table or list component for displaying memory usage details.

c) GPU Usage Section:
   - Similar to the Memory Usage section, use bar graphs for GPU usage and memory.
   - Implement a table or list component for GPU details.

d) Process Monitor Section:
   - Implement a sortable table component to display process information.

e) Overall Layout:
   - Use Bubble Tea's layout components (or create custom ones) to divide the screen into four main sections.
   - Ensure that each section can be updated independently for better performance.

2) Develop a basic Bubble Tea prototype displaying sample data.
   - Start by implementing the overall layout structure.
   - Add placeholder components for each section.
   - Gradually replace placeholders with functional components, starting with CPU usage.

3) Integrate real-time data collection:
   - Implement data collection routines using gopsutil and nvml.
   - Set up Go channels to pass system metrics updates to the Bubble Tea event loop.

4) Refine UI components:
   - Implement interactive features (e.g., sorting processes, focusing on specific CPU cores).
   - Add color schemes and styling to improve readability and visual appeal.

5) Optimize performance:
   - Ensure smooth updates even with rapidly changing data.
   - Implement efficient rendering techniques, especially for the CPU Core Heatmap.

This document will continue to evolve as we refine the architecture and implementation details.
