package main

import (
	"fmt"
)

func buildSSHCommand(server *Server, parentGroup *Group) string {
	sshBinary := "ssh"
	port := 22
	user := ""
	var extraArgs []string

	if parentGroup != nil {
		parentSettings := getInheritedGroupSettings(parentGroup)

		if parentSettings.SSHBinary != "" {
			sshBinary = parentSettings.SSHBinary
		}

		if parentSettings.Port != 0 {
			port = parentSettings.Port
		}

		if parentSettings.User != "" {
			user = parentSettings.User
		}

		if len(parentSettings.ExtraArgs) > 0 {
			extraArgs = append(extraArgs, parentSettings.ExtraArgs...)
		}
	}

	if server.SSHBinary != "" {
		sshBinary = server.SSHBinary
	}

	if server.Port != 0 {
		port = server.Port
	}

	if server.User != "" {
		user = server.User
	}

	var args []string

	if port != 22 {
		args = append(args, "-p", fmt.Sprintf("%d", port))
	}

	if len(extraArgs) > 0 {
		args = append(args, extraArgs...)
	}

	if len(server.ExtraArgs) > 0 {
		args = append(args, server.ExtraArgs...)
	}

	hostStr := server.Host
	if user != "" {
		hostStr = fmt.Sprintf("%s@%s", user, server.Host)
	}
	args = append(args, hostStr)

	cmdString := sshBinary
	for _, arg := range args {
		cmdString += " " + arg
	}

	return cmdString
}
