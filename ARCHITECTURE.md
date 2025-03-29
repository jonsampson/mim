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

## Next Steps
1. Define the initial **static UI layout** (grid structure for CPU, GPU, and memory monitoring).
2. Implement the **data collection pipeline** using `gopsutil`, `nvml`, and `go_amd_smi`.
3. Set up **Go channels** for efficient data updates.
4. Develop a **basic Bubble Tea prototype** displaying sample data.

This document will evolve as we refine the architecture and implementation details.

