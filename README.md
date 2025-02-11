# waitup

A simple utility to check when a system becomes available via RDP or SSH. Perfect for monitoring system reboots or waiting for a server to come online.

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
- By default, checks RDP (3389) and SSH (22)
- Use -p flag to monitor a specific port

Progress is shown with dots, and you'll get a notification when the system is ready.

## License

MIT License - See LICENSE file for details 
