package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

const helpText = `waitup - A tool to monitor system availability via SSH or RDP

Usage:
    waitup [OPTIONS] HOSTNAME|IP

Options:
    -p, --port PORT       Check if a system is available on a specific port
    -q, --quiet          Suppress normal output, only exit with status code
    -t, --timeout DURATION  Maximum time to wait before exiting (e.g., 30s, 5m, 1h)
    -h, --help           Show this help message
    -v, --version        Show version information

Examples:
    waitup server1.example.com                 Monitor server1.example.com (SSH/RDP)
    waitup 192.168.1.100                       Monitor IP address 192.168.1.100 (SSH/RDP)
    waitup server1 -p 8080                     Monitor specific port 8080
    waitup 10.0.0.1 -p 443                     Monitor specific port 443
    waitup server1 --quiet --timeout 30s       Wait up to 30 seconds silently
    waitup server1 -q -t 5m                    Wait up to 5 minutes silently

Exit Codes:
    0    Connection established successfully
    1    Timeout reached or error occurred

The program will continuously check the specified port(s) until one becomes available.
A dot will be displayed every 5 seconds while waiting (unless --quiet is used).
`

var version = "dev" // this will be set by goreleaser

type Config struct {
	host       string
	quiet      bool
	timeout    time.Duration
	ports      []string
	sshEnabled bool
	rdpEnabled bool
}

func main() {
	// Define flags
	var (
		portFlag    string
		quietFlag   bool
		timeoutFlag string
		helpFlag    bool
		versionFlag bool
	)

	flag.StringVar(&portFlag, "p", "", "Port to check")
	flag.StringVar(&portFlag, "port", "", "Port to check")
	flag.BoolVar(&quietFlag, "q", false, "Quiet mode - suppress output")
	flag.BoolVar(&quietFlag, "quiet", false, "Quiet mode - suppress output")
	flag.StringVar(&timeoutFlag, "t", "", "Timeout duration (e.g., 30s, 5m, 1h)")
	flag.StringVar(&timeoutFlag, "timeout", "", "Timeout duration (e.g., 30s, 5m, 1h)")
	flag.BoolVar(&helpFlag, "h", false, "Show help")
	flag.BoolVar(&helpFlag, "help", false, "Show help")
	flag.BoolVar(&versionFlag, "v", false, "Show version")
	flag.BoolVar(&versionFlag, "version", false, "Show version")

	flag.Usage = func() {
		fmt.Print(helpText)
	}

	flag.CommandLine.Init(os.Args[0], flag.ContinueOnError)
	flag.CommandLine.SetOutput(os.Stderr)

	var hostname string
	var flagArgs []string
	skipNext := false

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]

		if skipNext {
			flagArgs = append(flagArgs, arg)
			skipNext = false
			continue
		}

		if strings.HasPrefix(arg, "-") {
			flagArgs = append(flagArgs, arg)
			if arg == "-p" || arg == "--port" || arg == "-t" || arg == "--timeout" {
				skipNext = true
			}
		} else {
			if hostname != "" {
				fmt.Fprintln(os.Stderr, "Error: multiple hostnames provided")
				fmt.Fprintln(os.Stderr, "Try 'waitup --help' for more information")
				os.Exit(1)
			}
			hostname = arg
		}
	}

	if err := flag.CommandLine.Parse(flagArgs); err != nil {
		os.Exit(1)
	}

	if versionFlag {
		fmt.Printf("waitup version %s\n", version)
		fmt.Println("https://github.com/creaked/waitup")
		os.Exit(0)
	}

	if helpFlag {
		fmt.Print(helpText)
		os.Exit(0)
	}

	// Validate timeout BEFORE checking for hostname
	// This ensures flag validation errors are reported first
	var timeoutDuration time.Duration
	if timeoutFlag != "" {
		timeoutStr := timeoutFlag
		// First, try to parse as a duration (e.g., "30s", "5m")
		if _, err := time.ParseDuration(timeoutFlag); err != nil {
			// If that fails, check if it's a plain number (e.g., "30")
			// Use strconv.ParseFloat to ensure the ENTIRE string is a valid number
			if _, numErr := strconv.ParseFloat(timeoutFlag, 64); numErr == nil {
				timeoutStr = timeoutFlag + "s"
			}
			// If it's neither a valid duration nor a valid number, timeoutStr remains unchanged
			// and will fail in the next ParseDuration call
		}

		duration, err := time.ParseDuration(timeoutStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid timeout duration '%s': %v\n", timeoutFlag, err)
			fmt.Fprintln(os.Stderr, "Use format like: 30s, 5m, 1h, 90m, or just a number for seconds (e.g., 30)")
			os.Exit(1)
		}
		if duration <= 0 {
			fmt.Fprintln(os.Stderr, "Error: timeout must be positive")
			os.Exit(1)
		}
		timeoutDuration = duration
	}

	// Validate port number if provided
	if portFlag != "" {
		portNum, err := strconv.Atoi(portFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid port '%s': must be a number\n", portFlag)
			os.Exit(1)
		}
		if portNum < 1 || portNum > 65535 {
			fmt.Fprintf(os.Stderr, "Error: invalid port '%s': must be between 1 and 65535\n", portFlag)
			os.Exit(1)
		}
	}

	if hostname == "" {
		fmt.Fprintln(os.Stderr, "Error: hostname or IP address required")
		fmt.Fprintln(os.Stderr, "Try 'waitup --help' for more information")
		os.Exit(1)
	}

	config := Config{
		host:    hostname,
		quiet:   quietFlag,
		timeout: timeoutDuration,
	}

	if portFlag != "" {
		config.ports = []string{portFlag}
		if portFlag == "22" && isSSHAvailable() {
			config.sshEnabled = true
		}
		if portFlag == "3389" && isRDPAvailable() {
			config.rdpEnabled = true
		}
	} else {
		config.ports = []string{"3389", "22"}
		config.sshEnabled = isSSHAvailable()
		config.rdpEnabled = isRDPAvailable()
	}

	runMonitoring(config)
}

func runMonitoring(config Config) {
	interval := 5 * time.Second

	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen, color.Bold).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	portDesc := "SSH/RDP"
	if len(config.ports) == 1 {
		portDesc = fmt.Sprintf("port %s", config.ports[0])
	}

	if !config.quiet {
		fmt.Printf(">> %s %s (%s)",
			cyan("Waiting for"),
			green(config.host),
			yellow(portDesc))
	}

	attempts := 0
	startTime := time.Now()

	var timeoutChan <-chan time.Time
	if config.timeout > 0 {
		timeoutChan = time.After(config.timeout)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	if checkConnection(config, &attempts, startTime, cyan, green, yellow) {
		return
	}

	for {
		select {
		case <-timeoutChan:
			if !config.quiet {
				elapsed := time.Since(startTime).Round(time.Second)
				fmt.Printf("\n>> %s\n", yellow("Timeout reached"))
				fmt.Printf(">> %s: %s\n", cyan("Time elapsed"), yellow(elapsed))
				fmt.Printf(">> %s: %d\n", cyan("Total attempts"), attempts)
			}
			os.Exit(1)
		case <-ticker.C:
			if checkConnection(config, &attempts, startTime, cyan, green, yellow) {
				return
			}
		}
	}
}

func checkConnection(config Config, attempts *int, startTime time.Time, cyan, green, yellow func(...interface{}) string) bool {
	*attempts++
	for _, port := range config.ports {
		address := net.JoinHostPort(config.host, port)
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

			if !config.quiet {
				fmt.Printf("\n>> %s\n", green("Connection Established!"))
				fmt.Printf(">> %s: %s\n", cyan("System"), green(config.host))
				fmt.Printf(">> %s: %s (%s)\n", cyan("Available on"), green(port), yellow(service))
				fmt.Printf(">> %s: %s\n", cyan("Time elapsed"), yellow(elapsed))
				fmt.Printf(">> %s: %d\n", cyan("Total attempts"), *attempts)

				if service == "SSH" && config.sshEnabled {
					if promptYesNo("\nWould you like to connect via SSH? (y/N) ") {
						connectSSH(config.host)
					}
				} else if service == "RDP" && config.rdpEnabled {
					if promptYesNo("\nWould you like to connect via RDP? (y/N) ") {
						connectRDP(config.host)
					}
				}
			}
			os.Exit(0)
		}
	}

	if !config.quiet {
		dot := yellow(".")
		fmt.Print(dot)
	}
	return false
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
	case "windows":
		cmd := exec.Command("mstsc", "/v:"+host)
		err := cmd.Start()
		if err != nil {
			fmt.Printf("Error launching RDP: %v\n", err)
		}
	default:
		// Linux/Unix systems
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
		} else {
			fmt.Println("Error: No RDP client found. Please install xfreerdp or rdesktop.")
		}
	}
}
