package main

import (
	"bufio"
	"fmt"

	"github.com/melbahja/goph"
	"github.com/pkg/errors"
	"io"
	"log"
	"os"
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
	args := os.Args
	if len(args) < 3 {
		log.Fatal("Usage: git-remote-nostr <remoteName> <remoteUrl>")
	}

	logOut = openLogFile()
	defer logOut.Close()

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

func DoConnect(command, remote string) (err error) {
	// Connects to given git service.

	auth, err := goph.UseAgent()
	if err != nil {
		panic(err)
	}

	// Do something with auth
	client, err := goph.New("git", "github.com", auth)
	if err != nil {
		panic(err)
	}

	// Do something with client
	cmd, err := client.Command(command, fmt.Sprintf("'%s'", remote))
	if err != nil {
		panic(err)
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	os.Stdout.WriteString("\n")
	if err := cmd.Start(); err != nil {
		panic(err)
	}

	if err := cmd.Wait(); err != nil {
		panic(err)
	}
	return
}
