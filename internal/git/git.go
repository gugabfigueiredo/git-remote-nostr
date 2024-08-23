package git

import (
	"os/exec"
)

type config struct {
	UserName  string
	UserEmail string
}

var Config config

func ProcessConfig() {
	userName, _ := exec.Command("git", "config", "--get", "user.name").Output()
	userEmail, _ := exec.Command("git", "config", "--get", "user.email").Output()

	Config = config{
		UserName:  string(userName),
		UserEmail: string(userEmail),
	}
}
