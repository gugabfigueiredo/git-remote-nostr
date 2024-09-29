package nostr

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

func (s *Service) Helper(remote *domain.Remote) error {
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
			if err := s.doConnect("git-upload-pack", remote); err != nil {
				return err
			}
		case command == "connect git-receive-pack\n":
			if err := s.doConnect("git-receive-pack", remote); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown command: %s", command)
		}
	}
}

func (s *Service) doConnect(command string, remote *domain.Remote, options ...util.CmdOption) error {
	cmd := exec.Command("ssh", remote.Login(), command, remote.Path())
	return util.RunCMD(cmd, options...)
}
