package git

import (
	"fmt"
	"github.com/gugabfigueiredo/git-remote-nostr/internal/domain"
	"os"
	"os/exec"
)

type Config struct {
	UserName  string
	UserEmail string
}

func GetConfig() Config {
	userName, _ := exec.Command("git", "config", "--get", "user.name").Output()
	userEmail, _ := exec.Command("git", "config", "--get", "user.email").Output()

	return Config{
		UserName:  string(userName),
		UserEmail: string(userEmail),
	}
}

func Helper(remoteName string, remote *domain.Remote) error {
	cmd := exec.Command("git", "remote-"+remote.Protocol, remoteName, remote.String())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Println()
	return cmd.Run()
}
