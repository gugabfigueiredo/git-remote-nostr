package main

import (
	"bufio"
	"fmt"
	"github.com/melbahja/goph"
	"github.com/pkg/errors"
	"io"
	"log"
	"os"
	"strings"
)

var logOut *os.File

func openLogFile() *os.File {
	// Specify the log file path
	logFile := "application.log"

	// Attempt to open or create the log file
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	// Set the output of the logger to the file
	log.SetOutput(file)
	return file
}

func main() {
	logOut = openLogFile()
	defer logOut.Close()

	args := os.Args
	log.Println("Starting git-remote-nostr", args)
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
	// Creates ssh auth agent
	auth, err := goph.Key("/home/gugabfigueiredo/.ssh/id_ed25519", "")
	//auth, err := goph.UseAgent()
	if err != nil {
		return err
	}

	// Create ssh client
	client, err := goph.New("git", "github.com", auth)
	if err != nil {
		return err
	}

	remote = strings.TrimPrefix(remote, "git@github.com:")

	// Create command
	cmd, err := client.Command(command, fmt.Sprintf("'%s'", remote))
	if err != nil {
		return err
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println()
	log.Printf("Connecting with %s at %s\n", command, remote)
	if err := cmd.Start(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}
