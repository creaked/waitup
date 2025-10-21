# waitup
[![GitHub release](https://img.shields.io/github/v/release/creaked/waitup)](https://github.com/creaked/waitup/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/creaked/waitup)](https://goreportcard.com/report/github.com/creaked/waitup)
[![License](https://img.shields.io/github/license/creaked/waitup)](LICENSE)

A utility to check when a system becomes available via SSH or RDP. Perfect for monitoring system reboots or waiting for a server to come online.

![Demo of waitup in action](media/demo.gif)

## Features
- Monitors systems for SSH (22) or RDP (3389) availability
- Supports custom port monitoring
- Automatic SSH client detection and connection
- RDP support for Windows (mstsc), macOS (Microsoft Remote Desktop), and Linux (xfreerdp/rdesktop)
- Progress indicator while waiting
- Connection statistics (time elapsed, attempts)
- Quiet mode for automation and scripting (suppresses output, returns exit codes)
- Timeout support to prevent indefinite waiting

## Install

### Using Homebrew
```bash
brew install creaked/tap/waitup
```

### AUR
```bash
yay -S waitup-bin
```

### Manual Installation

Download the latest release for your platform from the [releases page](https://github.com/creaked/waitup/releases/latest)

### Build from Source
```console
# Clone the repository
git clone https://github.com/creaked/waitup.git
cd waitup

# Build the binary
go build
```

## Usage

### Basic Usage
```console
# Monitor a host (default: SSH/RDP ports)
waitup hostname

# Examples
waitup server1.example.com
waitup 192.168.1.100

# Monitor specific port
waitup -p 8080 server1
waitup -p 443 10.0.0.1
```

### Automation & Scripting
Perfect for CI/CD pipelines and automation scripts:

```bash
# Quiet mode - suppresses output, only returns exit code
waitup -q server1.example.com
waitup --quiet 192.168.1.100

# With timeout - prevents indefinite waiting
waitup -t 30s server1.example.com      # Wait up to 30 seconds
waitup --timeout 5m 192.168.1.100      # Wait up to 5 minutes

# Combined for scripting
waitup -q -t 2m -p 8080 server1        # Quiet mode with 2-minute timeout

# Use in scripts
if waitup -q -t 30s myserver.com; then
    echo "Server is ready!"
    # Deploy or connect
else
    echo "Server did not come up in time"
    exit 1
fi
```

### Options
- `-p, --port PORT` - Monitor a specific port
- `-q, --quiet` - Suppress output (useful for scripts)
- `-t, --timeout DURATION` - Maximum wait time (e.g., 30s, 5m, 1h)
- `-h, --help` - Show help message
- `-v, --version` - Show version information

### Exit Codes
- `0` - Connection established successfully
- `1` - Timeout reached or error occurred

### Behavior
The tool will continuously check the specified port(s) until one becomes available:
- By default, checks SSH (22) and RDP (3389)
- Use `-p` flag to monitor a specific port
- Progress is shown with dots (unless `--quiet` is used)

When a connection is established (in interactive mode):
- For SSH: Prompts to connect using your system's SSH client
- For RDP: Launches your system's default RDP client

## Requirements
- For SSH connections: SSH client installed
- For RDP connections:
  - Windows: Built-in Remote Desktop client (mstsc)
  - macOS: Microsoft Remote Desktop app
  - Linux: xfreerdp or rdesktop

## License

MIT License - See [LICENSE](LICENSE) file for details 
