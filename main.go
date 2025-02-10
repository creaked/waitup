package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/fatih/color"
)

const helpText = `
waitup - A tool to monitor system availability via RDP or SSH

Usage:
    waitup SYSTEM_NAME    Check if a system is available via RDP (3389) or SSH (22)
    waitup -h, --help     Show this help message

Examples:
    waitup server1.example.com    Monitor server1.example.com
    waitup 192.168.1.100         Monitor IP address 192.168.1.100

The program will continuously check both ports until one becomes available.
A dot will be displayed every 5 seconds while waiting.
`

func main() {
	if len(os.Args) != 2 {
		printUsageAndExit()
	}

	if os.Args[1] == "-h" || os.Args[1] == "--help" {
		fmt.Print(helpText)
		os.Exit(0)
	}

	host := os.Args[1]
	ports := []string{"3389", "22"}
	interval := 5 * time.Second

	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen, color.Bold).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	fmt.Printf("\n🔍 %s %s\n", 
		cyan("Waiting for RDP/SSH availability on"),
		green(host))
	
	attempts := 0
	startTime := time.Now()

	for {
		attempts++
		for _, port := range ports {
			address := fmt.Sprintf("%s:%s", host, port)
			conn, err := net.DialTimeout("tcp", address, 3*time.Second)
			
			if err == nil {
				conn.Close()
				service := "RDP"
				if port == "22" {
					service = "SSH"
				}
				
				elapsed := time.Since(startTime).Round(time.Second)
				fmt.Printf("\n✅ %s\n", green("Connection Established!"))
				fmt.Printf("🖥️  %s: %s\n", cyan("System"), green(host))
				fmt.Printf("🔌 %s: %s (%s)\n", cyan("Available on"), green(port), yellow(service))
				fmt.Printf("⏱️  %s: %s\n", cyan("Time elapsed"), yellow(elapsed))
				fmt.Printf("🔄 %s: %d\n\n", cyan("Total attempts"), attempts)
				os.Exit(0)
			}
		}

		dot := yellow(".")
		fmt.Print(dot)
		time.Sleep(interval)
	}
}

func printUsageAndExit() {
	fmt.Println("Usage: waitup SYSTEM_NAME")
	fmt.Println("Try 'waitup --help' for more information")
	os.Exit(1)
} 