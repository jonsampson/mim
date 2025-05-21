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
│   │             #   - `CPUProcessInfo` and `GPUProcessInfo` now include a `User` field.
│   ├── service/  # Business logic (e.g., MetricsCollector, ProcessAnalyzer)
│   ├── tui/      # UI components (Bubble Tea model, views, messages)
│   │             #   - `process_monitor.go` now displays the process `User` in its tables.
│   ├── infra/    # External dependencies (e.g., gopsutil, nvml, go_amd_smi wrappers)
│   │             #   - Collectors (`cpu_memory_collector.go`, `nvidia_gpu_collector.go`)
│   │             #     now fetch process user information, with a fallback to UID
│   │             #     for "unknown userid" errors.
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

1. Refine the Process Monitor:
   - **[Partially Addressed]** Display more process information (User field added).
   - Improve sorting and filtering options for processes (e.g., by user).
   - Enhance the symbol allocation algorithm for better distribution.

2. Optimize performance:
   - Implement more efficient data update mechanisms
   - Reduce unnecessary redraws in the UI

3. Enhance GPU monitoring:
   - Implement more detailed GPU metrics display
   - Add support for multiple GPUs

4. Improve test coverage:
   - Develop unit tests for the new ProcessMonitor and SymbolAllocator components
   - Implement integration tests for the entire TUI

5. Documentation:
   - Update user documentation to reflect new features and UI changes
   - Enhance code documentation, especially for the new components

Here's the proposed layout for the initial static UI:

```
+------------------------------------------------------------------+
|                     CPU & GPU Usage Over Time                    |
|                                                                  |
| [CPU & GPU Usage Line Graph]                                     |
| (Showing total CPU % and GPU % usage over time)                  |
|                                                                  |
+------------------------------------------------------------------+
|                        CPU Core Heatmap                          |
|                                                                  |
| +------------------------+       +------------------------+       |
| |                        |       |                        |       |
| |   [CPU Core Heatmap]   |       |   [CPU Core Heatmap]   |       |
| |   (First half of       |       |   (Second half of      |       |
| |    cores)              |       |    cores)              |       |
| |                        |       |                        |       |
| +------------------------+       +------------------------+       |
|                                                                  |
+------------------------------------------------------------------+
|                Memory & GPU Memory Usage Over Time               |
|                                                                  |
| [Memory & GPU Memory Line Graph]                                 |
| (Showing total RAM % and GPU memory % usage over time)           |
|                                                                  |
+------------------------------------------------------------------+
|                        Process Monitor                           |
|                                                                  |
| [Top Processes Table]                                            |
|                                                                  |
| PID    User        CPU%    MEM%    Command                       |
| 1234   jsampson    25.0    10.2    process1                      |
| 5678   root        15.5     8.7    process2                      |
| 9012   1000        10.2     5.3    process3 (uid fallback)       |
| 3456   another     8.7     4.1    process4                      |
| 7890   jsampson    7.3     3.8    process5                      |
|                                                                  |
+------------------------------------------------------------------+
|                         GPU Details                              |
|                                                                  |
| Model:        [GPU Model]                                        |
| Temperature:  [Temperature]                                      |
| Fan Speed:    [Fan Speed]                                        |
| Power Usage:  [Power Usage]                                      |
|                                                                  |
+------------------------------------------------------------------+
```

This document will continue to evolve as we refine the architecture and implementation details.
