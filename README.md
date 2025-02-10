# waitup

A simple utility to check when a system becomes available via RDP or SSH. Perfect for monitoring system reboots or waiting for a server to come online.

## Install

### Using Homebrew (macOS/Linux)
```bash
brew install creaked/tap/waitup
```

### Using Chocolatey (Windows)
```console
choco install waitup
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
```
The tool will continuously check ports 3389 (RDP) and 22 (SSH) until one becomes available. Progress is shown with dots, and you'll get a notification when the system is ready.

## License

MIT License - See LICENSE file for details 
