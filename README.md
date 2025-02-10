# waitup

A simple utility to check when a system becomes available via RDP or SSH. Perfect for monitoring system reboots or waiting for a server to come online.

## Install

    # Clone the repository
    git clone https://github.com/creaked/waitup.git
    cd waitup

    # Build the binary
    go build

## Usage

    # Basic usage
    waitup hostname

    # Examples
    waitup server1.example.com
    waitup 192.168.1.100

The tool will continuously check ports 3389 (RDP) and 22 (SSH) until one becomes available. Progress is shown with dots, and you'll get a notification when the system is ready.

## Build from source

Requirements:
- Go 1.21 or later

    git clone https://github.com/creaked/waitup.git
    cd waitup
    go build

## License

MIT License - See LICENSE file for details 