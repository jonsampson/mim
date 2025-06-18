# Mim - System Resource Monitor

**Mim (Monitoring is Mim)** is a terminal-based user interface (TUI) application designed to provide a comprehensive, single-screen view of your system's resources. It helps you monitor CPU usage (aggregate and per-core), GPU usage (currently NVIDIA, with plans for AMD), memory consumption (system and GPU), and running processes.

<!-- Placeholder for Hero Image/Screenshot -->
<!-- TODO: Replace this comment with an actual screenshot of Mim in action -->
**[Screenshot Placeholder: A visually appealing screenshot of the Mim TUI showcasing its various panels like CPU graphs, process list, and GPU stats.]**

## Features

*   **CPU Monitoring:**
    *   Total CPU usage percentage.
    *   Per-core CPU usage displayed as sparklines.
    *   CPU core heatmap for a visual overview of core utilization.
    *   Historical graph of total CPU usage over time.
*   **GPU Monitoring (NVIDIA):**
    *   GPU utilization percentage.
    *   GPU memory usage percentage.
    *   List of processes running on the GPU, including PID, user, SM utilization, and GPU memory used.
    *   Historical graph of GPU usage over time.
    *   (Support for AMD GPUs is planned for future releases).
*   **Memory Monitoring:**
    *   Total system memory usage percentage.
    *   Total GPU memory usage percentage (for NVIDIA GPUs).
    *   Historical graph of memory usage over time.
*   **Process Monitor:**
    *   Lists top CPU-consuming processes with PID, User, CPU %, Memory %, and Command.
    *   Lists top Memory-consuming processes with PID, User, CPU %, Memory %, and Command.
    *   Lists top GPU-consuming processes (NVIDIA) with PID, User, SM Util %, and GPU Memory, and Command.
*   **Cross-Platform:** Built with Go, aiming for wide compatibility (Linux, macOS, Windows with caveats for GPU monitoring).
*   **Dynamic resizing:** The TUI adapts to your terminal window size.

## Prerequisites

*   **Go:** Version 1.18 or higher is recommended.
*   **For NVIDIA GPU Monitoring:**
    *   NVIDIA drivers installed.
    *   NVML (NVIDIA Management Library) installed and accessible.

## Installation

### Option 1: Install directly with Go

```bash
go install github.com/jonsampson/mim/cmd/mim@latest
```

This will install the `mim` binary to your `$GOPATH/bin` directory (typically `~/go/bin`). Make sure this directory is in your `PATH`.

### Option 2: Build from source

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/jonsampson/mim.git
    cd mim
    ```

2.  **Build the application:**
    ```bash
    go build ./cmd/mim/...
    ```
    This will create an executable (e.g., `mim` or `mim.exe`) in your current directory or your Go bin path depending on your Go setup. For a specific output path:
    ```bash
    go build -o mim_binary ./cmd/mim/main.go
    ```

## Running the Application

If you installed with `go install`, simply run:

```bash
mim
```

If you built from source, run the executable:

```bash
./mim_binary
```
(Or simply `mim` if it's in your PATH).

## Usage (TUI Keybindings)

*   **`q` or `Ctrl+c`**: Quit the application.
*   **Arrow Keys (`↑`/`↓`) or `k`/`j`**: Scroll through scrollable views (like process lists if they become scrollable, or main content if it exceeds screen height).
*   **`PageUp` / `PageDown`**: Scroll up/down by half a page.
*   **`Home` / `End`**: Scroll to the top/bottom.

The TUI provides several panels:
*   **CPU & GPU Usage Graph:** Shows historical data for overall CPU and GPU utilization.
*   **CPU Combined View:** Includes CPU usage sparklines for each core and a CPU heatmap.
*   **Memory Usage Graph:** Shows historical data for system RAM and GPU memory utilization.
*   **Process Monitor:** Contains tables for top processes by CPU, Memory, GPU utilization, and GPU Memory.

## Architecture

Mim is built using Go and the Bubble Tea framework, following The Elm Architecture. For more detailed information on the internal design, components, and data flow, please see [ARCHITECTURE.md](ARCHITECTURE.md).

## Contributing

Contributions are welcome! If you'd like to contribute, please:

1.  Fork the repository.
2.  Create a new branch for your feature or bug fix.
3.  Make your changes.
4.  Add tests for your changes if applicable.
5.  Ensure your code lints and builds correctly.
6.  Submit a pull request with a clear description of your changes.

(Further details on development setup and guidelines can be added here or in a separate `CONTRIBUTING.md` file).

## License

This project is licensed under the GNU General Public License v3.0. See the [LICENSE](LICENSE) file for details.
