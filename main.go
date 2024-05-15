package main

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"time"
)

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

// parseRemoteURL parses the URL and returns the scheme if valid, otherwise an error.
func parseRemoteURL(remoteUrl string) (string, error) {
	if u, err := url.Parse(remoteUrl); err == nil && u.Scheme != "" {
		return u.Scheme, nil
	}

	// Check for SCP-like SSH URLs first because they don't have a "scheme://"
	scpRe := regexp.MustCompile(`^[a-zA-Z0-9._@/+~-]+@[a-zA-Z0-9._@/+~-]+:[a-zA-Z0-9._@/+~-]+(\.git)?$`)
	if scpRe.MatchString(remoteUrl) {
		return "ssh", nil
	}

	return "", errors.New("invalid Git URL scheme")
}

func Main() (er error) {

	time.Sleep(30 * time.Second)
	if len(os.Args) < 3 {
		return fmt.Errorf("Usage: git-remote-nostr remote-name remote-url")
	}

	remoteName := os.Args[1]
	remoteUrl := os.Args[2]
	log.Printf("running git-remote-nostr: %s:%s\n", remoteName, remoteUrl)

	scheme, err := parseRemoteURL(remoteUrl)
	if err != nil {
		return fmt.Errorf("Invalid remote URL %s: %v", remoteUrl, err)
	}

	switch scheme {
	case "ssh":
		return doSSH(remoteName, remoteUrl)
	default:
		return doRemoteHelper(remoteName, remoteUrl, scheme)
	}
}

func doRemoteHelper(remoteName, remoteUrl, scheme string) error {
	log.Printf("git %s %s %s", fmt.Sprintf("remote-%s", scheme), remoteName, remoteUrl)
	cmd := exec.Command("git", fmt.Sprintf("remote-%s", scheme), remoteName, remoteUrl)
	cmd.Env = append(os.Environ(), "GIT_TRACE_PACKET=1", "GIT_TRACE=1")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func main() {
	logOut := openLogFile()
	defer logOut.Close()

	if err := Main(); err != nil {
		log.Fatal(err)
	}
}
