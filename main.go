package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"time"
	"runtime"

	"github.com/fatih/color"
)

const helpText = `waitup - A tool to monitor system availability via SSH or RDP

Usage:
    waitup HOSTNAME|IP              Check if a system is available via SSH (22) or RDP (3389)
    waitup HOSTNAME|IP -p PORT      Check if a system is available on a specific port
    waitup -h, --help              Show this help message
    waitup -v, --version           Show version information

Examples:
    waitup server1.example.com     Monitor server1.example.com (SSH/RDP)
    waitup 192.168.1.100          Monitor IP address 192.168.1.100 (SSH/RDP)
    waitup server1 -p 8080        Monitor specific port 8080
    waitup 10.0.0.1 -p 443       Monitor specific port 443

The program will continuously check the specified port(s) until one becomes available.
A dot will be displayed every 5 seconds while waiting.
`

var version = "dev" // this will be set by goreleaser

func main() {
	if len(os.Args) == 2 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		fmt.Printf("waitup version %s\n", version)
		fmt.Println("https://github.com/creaked/waitup")
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
	var sshEnabled, rdpEnabled bool

	if len(os.Args) == 4 && os.Args[2] == "-p" {
		ports = []string{os.Args[3]}
		if os.Args[3] == "22" && isSSHAvailable() {
			sshEnabled = true
		}
		if os.Args[3] == "3389" && isRDPAvailable() {
			rdpEnabled = true
		}
	} else if len(os.Args) == 2 {
		ports = []string{"3389", "22"}
		sshEnabled = isSSHAvailable()
		rdpEnabled = isRDPAvailable()
	} else {
		printUsageAndExit()
	}

	interval := 5 * time.Second

	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen, color.Bold).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	portDesc := "SSH/RDP"
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

				if service == "SSH" && sshEnabled {
					if promptYesNo("\nWould you like to connect via SSH? (y/N) ") {
						connectSSH(host)
					}
				} else if service == "RDP" && rdpEnabled {
					if promptYesNo("\nWould you like to connect via RDP? (y/N) ") {
						connectRDP(host)
					}
				}
				os.Exit(0)
			}
		}

		dot := yellow(".")
		fmt.Print(dot)
		time.Sleep(interval)
	}
}

func promptYesNo(prompt string) bool {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

func promptUsername(defaultUser string) string {
	fmt.Printf("Enter username (press Enter for %s): ", defaultUser)
	reader := bufio.NewReader(os.Stdin)
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	if username == "" {
		return defaultUser
	}
	return username
}

func printUsageAndExit() {
	fmt.Println("Usage: waitup HOSTNAME|IP [-p PORT]")
	fmt.Println("Try 'waitup --help' for more information")
	os.Exit(1)
}

func isSSHAvailable() bool {
	_, err := exec.LookPath("ssh")
	return err == nil
}

func connectSSH(host string) {
	currentUser, err := user.Current()
	if err != nil {
		fmt.Printf("Error getting current user: %v\n", err)
		return
	}

	username := promptUsername(currentUser.Username)
	sshHost := fmt.Sprintf("%s@%s", username, host)

	cmd := exec.Command("ssh", sshHost)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		fmt.Printf("Error connecting: %v\n", err)
	}
}

func isRDPAvailable() bool {
	switch runtime.GOOS {
	case "windows":
		_, err := exec.LookPath("mstsc")
		return err == nil
	case "darwin":
		_, err := os.Stat("/Applications/Microsoft Remote Desktop.app")
		return err == nil
	default:
		clients := []string{"xfreerdp", "rdesktop"}
		for _, client := range clients {
			if _, err := exec.LookPath(client); err == nil {
				return true
			}
		}
		return false
	}
}

func connectRDP(host string) {
	switch runtime.GOOS {
	case "darwin":
		rdpContent := fmt.Sprintf(`full address:s:%s`, host)
		
		// create a temp rdp file as a work around for macOS
		tmpFile, err := os.CreateTemp("", "waitup-*.rdp")
		if err != nil {
			fmt.Printf("Error creating RDP file: %v\n", err)
			return
		}
		defer os.Remove(tmpFile.Name())

		if _, err := tmpFile.WriteString(rdpContent); err != nil {
			fmt.Printf("Error writing RDP file: %v\n", err)
			return
		}
		tmpFile.Close()

		cmd := exec.Command("open", tmpFile.Name())
		err = cmd.Start()
		if err != nil {
			fmt.Printf("Error launching RDP: %v\n", err)
		}
		time.Sleep(time.Second)
	default:
		if _, err := exec.LookPath("xfreerdp"); err == nil {
			cmd := exec.Command("xfreerdp", "/v:"+host)
			err = cmd.Start()
			if err != nil {
				fmt.Printf("Error launching RDP: %v\n", err)
			}
		} else if _, err := exec.LookPath("rdesktop"); err == nil {
			cmd := exec.Command("rdesktop", host)
			err = cmd.Start()
			if err != nil {
				fmt.Printf("Error launching RDP: %v\n", err)
			}
		}
	}
} 