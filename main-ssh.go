package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func mainSSH() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		args := strings.Fields(scanner.Text())
		if len(args) < 1 {
			continue
		}

		command := args[0]
		switch command {
		case "capabilities":
			fmt.Println("fetch")
			fmt.Println("push")
			fmt.Println()
		case "list":
			if len(args) > 1 {
				list(args[1]) // Assuming second argument is the URL
			}
			fmt.Println()
		case "fetch":
			if len(args) > 2 {
				fetch(args[1], args[2], args[3]) // URL, SHA, ref
			}
			fmt.Println()
		case "push":
			if len(args) > 2 {
				push(args[1], args[2:]) // URL, [localRef:remoteRef, ...]
			}
			fmt.Println()
		}
	}
}

func executeSSHCommand(url, command string) {
	// Parse the URL to extract the host and path
	parts := strings.Split(url, "@")
	host := parts[1]
	path := parts[0]

	// SSH command construction
	sshCmd := exec.Command("ssh", host, command)
	sshCmd.Dir = path
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr
	if err := sshCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "SSH command failed: %v\n", err)
	}
}

func list(url string) {
	executeSSHCommand(url, "git-upload-pack --advertise-refs .")
}

func fetch(url, sha, ref string) {
	executeSSHCommand(url, fmt.Sprintf("git-upload-pack . '%s'", ref))
}

func push(url string, refs []string) {
	refSpecs := strings.Join(refs, " ")
	executeSSHCommand(url, fmt.Sprintf("git-receive-pack '%s'", refSpecs))
}
