package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
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
		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/bash"
		}

		var shellCommand string
		cmdParts := append([]string{sshBinary}, args...)
		sshCmd := strings.Join(cmdParts, " ")

		if sshBinary != "ssh" {
			switch {
			case strings.Contains(shell, "bash"):
				shellCommand = fmt.Sprintf("source ~/.bashrc > /dev/null 2>&1 || true; source ~/.bash_profile > /dev/null 2>&1 || true; exec %s", sshCmd)
			case strings.Contains(shell, "zsh"):
				shellCommand = fmt.Sprintf("source ~/.zshrc > /dev/null 2>&1 || true; exec %s", sshCmd)
			case strings.Contains(shell, "fish"):
				shellCommand = fmt.Sprintf("source ~/.config/fish/config.fish > /dev/null 2>&1 || true; exec %s", sshCmd)
			default:
				shellCommand = fmt.Sprintf("exec %s", sshCmd)
			}
			return fmt.Sprintf("SASHA_PASSWORD='%s' %s -i -c \"%s\"", password, shell, shellCommand)
		} else {
			shellCommand = fmt.Sprintf("exec %s", sshCmd)
			return fmt.Sprintf("SASHA_PASSWORD='%s' %s -c \"%s\"", password, shell, shellCommand)
		}
	}

	cmdString := sshBinary
	for _, arg := range args {
		cmdString += " " + arg
	}

	return cmdString
}

func extractHostFromCommand(cmdString string) string {
	if strings.Contains(cmdString, "SASHA_PASSWORD=") {
		sshCmdStart := strings.Index(cmdString, "exec ssh")
		if sshCmdStart == -1 {
			return ""
		}

		sshCmdFull := cmdString[sshCmdStart+5:]
		endQuote := strings.LastIndex(sshCmdFull, "\"")
		if endQuote != -1 {
			sshCmdFull = sshCmdFull[:endQuote]
		}

		parts := strings.Fields(strings.TrimSpace(sshCmdFull))
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
			executeSSHWithPasswordCommand(m.sshCommand)
		} else {
			executeNormalSSHCommand(m.sshCommand)
		}
	}
}

func executeSSHWithPasswordCommand(cmdString string) {
	parts := strings.SplitN(cmdString, " ", 2)
	if len(parts) < 2 || !strings.HasPrefix(parts[0], "SASHA_PASSWORD=") {
		fmt.Fprintf(os.Stderr, "Error: Invalid password command format\n")
		os.Exit(1)
	}

	passwordPart := parts[0]
	password := strings.TrimPrefix(passwordPart, "SASHA_PASSWORD='")
	password = strings.TrimSuffix(password, "'")

	remainingCmd := parts[1]

	sshCmdStart := strings.Index(remainingCmd, "exec ssh")
	if sshCmdStart == -1 {
		fmt.Fprintf(os.Stderr, "Error: Could not find SSH command\n")
		os.Exit(1)
	}

	sshCmdFull := remainingCmd[sshCmdStart+5:]

	endQuote := strings.LastIndex(sshCmdFull, "\"")
	if endQuote != -1 {
		sshCmdFull = sshCmdFull[:endQuote]
	}

	sshParts := parseSSHCommand(sshCmdFull)

	if len(sshParts) < 2 {
		fmt.Fprintf(os.Stderr, "Error: Invalid SSH command\n")
		os.Exit(1)
	}

	binary := sshParts[0]
	args := sshParts[1:]

	err := executeSSHWithPassword(binary, args, password)
	if err != nil {
		os.Exit(1)
	}
}

func parseSSHCommand(cmdStr string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false
	escaped := false

	cmdStr = strings.TrimSpace(cmdStr)

	for _, char := range cmdStr {
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

func executeNormalSSHCommand(cmdString string) {
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
