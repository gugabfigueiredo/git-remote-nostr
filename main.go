package main

import (
	"github.com/gugabfigueiredo/git-remote-nostr/internal/git"
	"github.com/gugabfigueiredo/git-remote-nostr/internal/nostr"
	"github.com/gugabfigueiredo/git-remote-nostr/internal/ssh"
	"log"
	"os"
)

func main() {
	args := os.Args
	if len(args) < 3 {
		log.Fatal("Usage: git-remote-nostr <remoteName> <remoteUrl>")
	}

	gitConfig := git.GetConfig()

	pvtk, pubk := nostr.ResolveCredentials(gitConfig.UserName, gitConfig.UserEmail)

	nostrClient := nostr.NewClient(pvtk, pubk)
	nostrService := nostr.NewService(nostrClient)

	remote, err := nostrService.ResolveRemote(args[2])
	if err != nil {
		log.Fatalf("Error resolving remote: %v: %s", err, args[2])
	}

	switch remote.Protocol {
	case "nostr":
		if err = nostrService.Helper(remote); err != nil {
			log.Fatalf("failed to setup remote helper: %v", err)
		}
	case "ssh":
		if err = ssh.Helper(remote); err != nil {
			log.Fatalf("failed to setup remote helper: %v", err)
		}
	case "http", "https", "git":
		if err = git.Helper(args[1], remote); err != nil {
			log.Fatalf("failed to setup remote helper: %v", err)
		}
	default:
		log.Fatalf("unsupported protocol for remote url: %s", remote.String())
	}
}
