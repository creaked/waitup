package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/fatih/color"
)

const helpText = `waitup - A tool to monitor system availability via RDP or SSH

Usage:
    waitup HOSTNAME|IP              Check if a system is available via RDP (3389) or SSH (22)
    waitup HOSTNAME|IP -p PORT      Check if a system is available on a specific port
    waitup -h, --help              Show this help message
    waitup -v, --version           Show version information

Examples:
    waitup server1.example.com     Monitor server1.example.com (RDP/SSH)
    waitup 192.168.1.100          Monitor IP address 192.168.1.100 (RDP/SSH)
    waitup server1 -p 8080        Monitor specific port 8080
    waitup 10.0.0.1 -p 443       Monitor specific port 443

The program will continuously check the specified port(s) until one becomes available.
A dot will be displayed every 5 seconds while waiting.
`

var version = "dev" // this will be set by goreleaser

func main() {
	if len(os.Args) == 2 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		fmt.Printf("waitup version %s\n", version)
		os.Exit(0)
	}

	if len(os.Args) == 2 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		fmt.Print(helpText)
		os.Exit(0)
	}

	if len(os.Args) < 2 || len(os.Args) > 4 {
		printUsageAndExit()
	}

	host := os.Args[1]
	var ports []string

	if len(os.Args) == 4 && os.Args[2] == "-p" {
		ports = []string{os.Args[3]}
	} else if len(os.Args) == 2 {
		ports = []string{"3389", "22"}
	} else {
		printUsageAndExit()
	}

	interval := 5 * time.Second

	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen, color.Bold).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	portDesc := "RDP/SSH"
	if len(ports) == 1 {
		portDesc = fmt.Sprintf("port %s", ports[0])
	}

	fmt.Printf(">> %s %s (%s)", 
		cyan("Waiting for"),
		green(host),
		yellow(portDesc))
	
	attempts := 0
	startTime := time.Now()

	for {
		attempts++
		for _, port := range ports {
			address := fmt.Sprintf("%s:%s", host, port)
			conn, err := net.DialTimeout("tcp", address, 3*time.Second)
			
			if err == nil {
				conn.Close()
				service := "Custom Port"
				if port == "3389" {
					service = "RDP"
				} else if port == "22" {
					service = "SSH"
				}
				
				elapsed := time.Since(startTime).Round(time.Second)
				fmt.Printf("\n>> %s\n", green("Connection Established!"))
				fmt.Printf(">> %s: %s\n", cyan("System"), green(host))
				fmt.Printf(">> %s: %s (%s)\n", cyan("Available on"), green(port), yellow(service))
				fmt.Printf(">> %s: %s\n", cyan("Time elapsed"), yellow(elapsed))
				fmt.Printf(">> %s: %d\n", cyan("Total attempts"), attempts)
				os.Exit(0)
			}
		}

		dot := yellow(".")
		fmt.Print(dot)
		time.Sleep(interval)
	}
}

func printUsageAndExit() {
	fmt.Println("Usage: waitup HOSTNAME|IP [-p PORT]")
	fmt.Println("Try 'waitup --help' for more information")
	os.Exit(1)
} 