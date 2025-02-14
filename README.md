# waitup

A utility to check when a system becomes available via SSH or RDP. Perfect for monitoring system reboots or waiting for a server to come online.

## Features
- Monitors systems for SSH (22) or RDP (3389) availability
- Supports custom port monitoring
- Automatic SSH client detection and connection
- RDP support for Windows (mstsc), macOS (Microsoft Remote Desktop), and Linux (xfreerdp/rdesktop)
- Progress indicator while waiting
- Connection statistics (time elapsed, attempts)

## Install

### Using Homebrew (macOS/Linux)
```bash
brew install creaked/tap/waitup
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
```console
# Basic usage
waitup hostname

# Examples
waitup server1.example.com
waitup 192.168.1.100

# Monitor specific port
waitup server1 -p 8080
waitup 10.0.0.1 -p 443
```
The tool will continuously check the specified port(s) until one becomes available:
- By default, checks SSH (22) and RDP (3389)
- Use -p flag to monitor a specific port

When a connection is established:
- For SSH: Prompts to connect using your system's SSH client
- For RDP: Launches your system's default RDP client

Progress is shown with dots, and you'll get a notification when the system is ready.

## Requirements
- For SSH connections: SSH client installed
- For RDP connections:
  - Windows: Built-in Remote Desktop client (mstsc)
  - macOS: Microsoft Remote Desktop app
  - Linux: xfreerdp or rdesktop

## License

MIT License - See [LICENSE](LICENSE) file for details 
