package main

import (
	"bufio"
	"fmt"
	"github.com/go-git/go-git/plumbing/transport"
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

	remote, err := transport.NewEndpoint(args[2])
	if err != nil {
		log.Fatal(err)
	}

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
			if err = DoConnect("git-upload-pack", remote); err != nil {
				log.Fatal(err)
			}
		case command == "connect git-receive-pack\n":
			if err = DoConnect("git-receive-pack", remote); err != nil {
				log.Fatal(err)
			}
		default:
			log.Fatalf("Unknown command: %q", command)
		}
	}
}

func DoConnect(command string, remote *transport.Endpoint) error {
	cmd := exec.Command("ssh", getRemoteLogin(remote), command, getRemotePath(remote))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Println()
	return cmd.Run()
}

func getRemoteLogin(remote *transport.Endpoint) string {
	hostAndPort := remote.Host
	port := remote.Port
	if port != 22 {
		hostAndPort = fmt.Sprintf("%s:%d", hostAndPort, port)
	}
	return remote.User + "@" + hostAndPort
}

func getRemotePath(remote *transport.Endpoint) string {
	return strings.TrimPrefix(remote.Path, "/")
}
