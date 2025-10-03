package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/gaetanlhf/sasha/tty"
)

func buildSSHCommand(server *Server, parentGroup *Group) string {
	sshBinary := "ssh"
	port := 22
	user := ""
	password := ""
	var extraArgs []string

	if parentGroup != nil {
		parentSettings := getInheritedGroupSettings(parentGroup)

		if parentSettings.SSHBinary != nil && *parentSettings.SSHBinary != "" {
			sshBinary = *parentSettings.SSHBinary
		}
		if parentSettings.Port != nil && *parentSettings.Port != 0 {
			port = *parentSettings.Port
		}
		if parentSettings.User != nil && *parentSettings.User != "" {
			user = *parentSettings.User
		}
		if parentSettings.Password != nil && *parentSettings.Password != "" {
			password = *parentSettings.Password
		}
		if len(parentSettings.ExtraArgs) > 0 {
			extraArgs = append([]string{}, parentSettings.ExtraArgs...)
		}
	}

	if server.SSHBinary != nil && *server.SSHBinary != "" {
		sshBinary = *server.SSHBinary
	}
	if server.Port != nil && *server.Port != 0 {
		port = *server.Port
	}
	if server.User != nil && *server.User != "" {
		user = *server.User
	}
	if server.Password != nil && *server.Password != "" {
		password = *server.Password
	}

	if len(server.ExtraArgs) > 0 {
		extraArgs = append([]string{}, server.ExtraArgs...)
	}

	var args []string

	if port != 22 {
		args = append(args, "-p", fmt.Sprintf("%d", port))
	}

	if len(extraArgs) > 0 {
		args = append(args, extraArgs...)
	}

	hostStr := server.Host
	if user != "" {
		hostStr = fmt.Sprintf("%s@%s", user, server.Host)
	}
	args = append(args, hostStr)

	if password != "" {
		return fmt.Sprintf("SASHA_PASSWORD='%s' SASHA_BINARY='%s' SASHA_ARGS='%s'",
			password, sshBinary, strings.Join(args, " "))
	}

	cmdString := sshBinary
	for _, arg := range args {
		cmdString += " " + arg
	}

	return cmdString
}

func extractHostFromCommand(cmdString string) string {
	if strings.Contains(cmdString, "SASHA_PASSWORD=") {
		argsStart := strings.Index(cmdString, "SASHA_ARGS='")
		if argsStart == -1 {
			return ""
		}
		argsStart += 12
		argsEnd := strings.Index(cmdString[argsStart:], "'")
		if argsEnd == -1 {
			return ""
		}
		argsStr := cmdString[argsStart : argsStart+argsEnd]
		parts := strings.Fields(argsStr)
		if len(parts) > 0 {
			host := parts[len(parts)-1]
			if strings.Contains(host, "@") {
				hostParts := strings.SplitN(host, "@", 2)
				if len(hostParts) == 2 {
					return hostParts[1]
				}
			}
			return host
		}
	} else {
		parts := strings.Fields(strings.TrimSpace(cmdString))
		if len(parts) > 0 {
			host := parts[len(parts)-1]
			if strings.Contains(host, "@") {
				hostParts := strings.SplitN(host, "@", 2)
				if len(hostParts) == 2 {
					return hostParts[1]
				}
			}
			return host
		}
	}
	return ""
}

func handleApplicationExit(finalModel interface{}) {
	if m, ok := finalModel.(model); ok && m.quitting && m.sshCommand != "" {
		host := extractHostFromCommand(m.sshCommand)
		if host != "" {
			fmt.Printf("Connecting to %s...\n", host)
		}

		if strings.Contains(m.sshCommand, "SASHA_PASSWORD=") {
			executeSSHWithPassword(m.sshCommand)
		} else {
			executeNormalSSH(m.sshCommand)
		}
	}
}

func executeSSHWithPassword(cmdString string) {
	var password, binary, argsStr string

	parts := strings.Split(cmdString, " ")
	for _, part := range parts {
		if strings.HasPrefix(part, "SASHA_PASSWORD='") {
			password = strings.TrimPrefix(part, "SASHA_PASSWORD='")
			password = strings.TrimSuffix(password, "'")
		}
		if strings.HasPrefix(part, "SASHA_BINARY='") {
			binary = strings.TrimPrefix(part, "SASHA_BINARY='")
			binary = strings.TrimSuffix(binary, "'")
		}
		if strings.HasPrefix(part, "SASHA_ARGS='") {
			idx := strings.Index(cmdString, "SASHA_ARGS='")
			if idx != -1 {
				argsStr = cmdString[idx+12:]
				endIdx := strings.LastIndex(argsStr, "'")
				if endIdx != -1 {
					argsStr = argsStr[:endIdx]
				}
			}
		}
	}

	if binary == "" || argsStr == "" {
		fmt.Fprintf(os.Stderr, "Error: Invalid command format\n")
		os.Exit(1)
	}

	binaryPath, err := exec.LookPath(binary)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: SSH binary not found: %v\n", err)
		os.Exit(1)
	}

	args := parseSSHArgs(argsStr)

	manager := tty.NewManager(password)
	if err := manager.Run(binaryPath, args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func parseSSHArgs(argsStr string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false
	escaped := false

	argsStr = strings.TrimSpace(argsStr)

	for _, char := range argsStr {
		if escaped {
			current.WriteRune(char)
			escaped = false
			continue
		}

		if char == '\\' {
			escaped = true
			continue
		}

		if char == '"' || char == '\'' {
			inQuotes = !inQuotes
			continue
		}

		if char == ' ' && !inQuotes {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
			continue
		}

		current.WriteRune(char)
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

func executeNormalSSH(cmdString string) {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
	}

	isCustomSSH := !strings.HasPrefix(strings.TrimSpace(cmdString), "ssh ")

	var cmd *exec.Cmd
	if isCustomSSH {
		var shellCommand string
		switch {
		case strings.Contains(shell, "bash"):
			shellCommand = fmt.Sprintf("source ~/.bashrc > /dev/null 2>&1 || true; source ~/.bash_profile > /dev/null 2>&1 || true; %s", cmdString)
		case strings.Contains(shell, "zsh"):
			shellCommand = fmt.Sprintf("source ~/.zshrc > /dev/null 2>&1 || true; %s", cmdString)
		case strings.Contains(shell, "fish"):
			shellCommand = fmt.Sprintf("source ~/.config/fish/config.fish > /dev/null 2>&1 || true; %s", cmdString)
		default:
			shellCommand = cmdString
		}
		cmd = exec.Command(shell, "-i", "-c", shellCommand)
	} else {
		cmd = exec.Command(shell, "-c", cmdString)
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
