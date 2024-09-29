package ssh

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/gugabfigueiredo/git-remote-nostr/internal/domain"
	"github.com/gugabfigueiredo/git-remote-nostr/internal/util"
	"io"
	"os"
	"os/exec"
)

func Helper(remote *domain.Remote) error {
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
	return util.RunCMD(cmd)
}
