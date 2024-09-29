package git

import (
	"github.com/gugabfigueiredo/git-remote-nostr/internal/domain"
	"github.com/gugabfigueiredo/git-remote-nostr/internal/util"
	"os/exec"
)

type config struct {
	UserName  string
	UserEmail string
}

var Config config

func Init() {
	userName, _ := exec.Command("git", "config", "--get", "user.name").Output()
	userEmail, _ := exec.Command("git", "config", "--get", "user.email").Output()

	Config = config{
		UserName:  string(userName),
		UserEmail: string(userEmail),
	}
}

func Helper(remoteName string, remote *domain.Remote) error {
	cmd := exec.Command("git", "remote-"+remote.Protocol, remoteName, remote.String())
	return util.RunCMD(cmd)
}
