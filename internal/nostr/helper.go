package nostr

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/gugabfigueiredo/git-remote-nostr/internal/domain"
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
			if err := s.doUploadPack(remote); err != nil {
				return err
			}
		case command == "connect git-receive-pack\n":
			if err := s.doReceivePack(remote); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown command: %s", command)
		}
	}
}

func (s *Service) doUploadPack(remote *domain.Remote) error {
	cmd := exec.Command("ssh", "git-upload-pack", remote.Path())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Println()
	return cmd.Run()
}

func (s *Service) doReceivePack(remote *domain.Remote) error {
	cmd := exec.Command("ssh", "git-receive-pack", remote.Path())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Println()
	return cmd.Run()
}
