package main

import (
	"bufio"
	"fmt"
	"github.com/melbahja/goph"
	"os"
	"strings"

	golog "github.com/ipfs/go-log/v2"
)

var logger = golog.Logger("remote-helper")

func main() {
	args := os.Args
	if len(args) < 3 {
		logger.Fatal("Usage: git-remote-nostr <remoteName>")
	}

	stdinReader := bufio.NewReader(os.Stdin)
	for {
		command, err := stdinReader.ReadString('\n')
		strings.Fields(command)
		if err != nil {
			logger.Fatal(err)
		}

		switch {
		case command == "capabilities\n":
			fmt.Fprintln(os.Stdout, "connect")
			fmt.Fprintln(os.Stdout, "")
		case command == "connect git-upload-pack\n":
			if err = DoConnect("git-upload-pack", ""); err != nil {
				logger.Fatal(err)
			}
		case command == "connect git-receive-pack\n":
			if err = DoConnect("git-receive-pack", ""); err != nil {
				logger.Fatal(err)
			}
		default:
			logger.Fatalf("Unknown command: %q", command)
		}
	}
}

func DoConnect(command, remote string) (err error) {
	// Connects to given git service.

	auth, err := goph.UseAgent()
	if err != nil {
		panic(err)
	}

	// Do something with auth
	client, err := goph.New("git", "github.com", auth)
	if err != nil {
		panic(err)
	}

	// Do something with client
	cmd, err := client.Command(command, "'gugabfigueiredo/git-remote-nostr.git'")
	if err != nil {
		panic(err)
	}

	os.Stdout.WriteString("\n")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	if err := cmd.Start(); err != nil {
		panic(err)
	}

	if err := cmd.Wait(); err != nil {
		panic(err)
	}
	return
}
