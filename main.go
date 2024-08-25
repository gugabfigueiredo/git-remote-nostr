package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/gugabfigueiredo/git-remote-nostr/internal/domain"
	"github.com/gugabfigueiredo/git-remote-nostr/internal/git"
	"github.com/gugabfigueiredo/git-remote-nostr/internal/nostr"
	"io"
	"log"
	"os"
	"os/exec"
)

var nostrService domain.RemoteService

func init() {
	git.ProcessConfig()
	pvtk, pubk := nostr.GetCredentials(git.Config.UserName, git.Config.UserEmail)
	nostrClient := nostr.NewClient(pvtk, pubk)
	nostrService = nostr.NewService(nostrClient)
}

func main() {

	args := os.Args
	if len(args) < 3 {
		log.Fatal("Usage: git-remote-nostr <remoteName> <remoteUrl>")
	}

	remote, err := nostrService.ResolveRemote(args[2])
	if err != nil {
		log.Fatalf("Error resolving remote: %v: %s", err, args[2])
	}

	switch remote.Protocol {
	case "ssh", "nostr":
		if err = nostrHelper(remote); err != nil {
			log.Fatalf("failed to setup remote helper: %v", err)
		}
	case "http", "https", "git":
		if err = defaultHelper(args[1], remote); err != nil {
			log.Fatalf("failed to setup remote helper: %v", err)
		}
	default:
		log.Fatalf("unsupported protocol for remote url: %s", remote.String())
	}
}

func doCMD(cmd *exec.Cmd) error {
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Println()
	return cmd.Run()
}

func defaultHelper(remoteName string, remote *domain.Remote) error {
	cmd := exec.Command("git", "remote-"+remote.Protocol, remoteName, remote.String())
	return doCMD(cmd)
}

func nostrHelper(remote *domain.Remote) error {
	stdinReader := bufio.NewReader(os.Stdin)
	for {
		command, err := stdinReader.ReadString('\n')
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return errors.Join(err, fmt.Errorf("failed to read command"))
		}

		switch {
		case command == "capabilities\n":
			fmt.Println("connect")
			fmt.Println()
		case command == "connect git-upload-pack\n":
			if err := doConnect("git-upload-pack", remote); err != nil {
				return err
			}
		case command == "connect git-receive-pack\n":
			if err := nostrService.Auth(remote); err != nil {
				return fmt.Errorf("failed to request upload: %v", err)
			}
			if err := doConnect("git-receive-pack", remote); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown command: %s", command)
		}
	}
}

func doConnect(command string, remote *domain.Remote) error {
	cmd := exec.Command("ssh", remote.Login(), command, remote.Path())
	return doCMD(cmd)
}
