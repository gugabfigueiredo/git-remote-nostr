package main

import (
	"bufio"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

func main() {
	args := os.Args
	if len(args) < 3 {
		log.Fatal("Usage: git-remote-nostr <remoteName> <remoteUrl>")
	}

	remoteUrl := args[2]

	stdinReader := bufio.NewReader(os.Stdin)
	for {
		command, err := stdinReader.ReadString('\n')
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		switch {
		case command == "capabilities\n":
			fmt.Println("connect")
			fmt.Println()
		case command == "connect git-upload-pack\n":
			if err = DoConnect("git-upload-pack", remoteUrl); err != nil {
				log.Fatal(err)
			}
		case command == "connect git-receive-pack\n":
			if err = DoConnect("git-receive-pack", remoteUrl); err != nil {
				log.Fatal(err)
			}
		default:
			log.Fatalf("Unknown command: %q", command)
		}
	}
}

func DoConnect(command, remote string) error {
	remote = strings.TrimPrefix(remote, "git@github.com:")
	cmd := exec.Command("ssh", "git@github.com", command, fmt.Sprintf("'%s'", remote))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Println()
	return cmd.Run()
}
