package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
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
		return doHTTP(remoteName, remoteUrl, scheme)
	}
}

func doSSH(remoteName, remoteUrl string) error {
	stdinReader := bufio.NewReader(os.Stdin)
	for {
		command, err := stdinReader.ReadString('\n')
		if err != nil {
			return err
		}
		log.Printf("ssh command: %q\n", command)

		switch {
		case command == "capabilities\n":
			fmt.Println("fetch")
			fmt.Println("push")
			fmt.Println()
		case command == "list\n":
			if err := handleList(remoteUrl); err != nil {
				return err
			}
		case strings.HasPrefix(command, "fetch"):
			args := strings.TrimPrefix(command, "fetch ")
			if err := handleFetch(remoteUrl, args); err != nil {
				return err
			}
		case strings.HasPrefix(command, "push"):
			if err := handlePush(remoteUrl); err != nil {
				return err
			}
		case command == "\n":
			fmt.Println()
		default:
			return fmt.Errorf("unknown command: %q", command)
		}
	}
}

func handleList(remoteUrl string) error {
	log.Println("git ls-remote", remoteUrl)
	cmd := exec.Command("git", "ls-remote", remoteUrl)
	cmd.Env = append(os.Environ(), "GIT_TRACE_PACKET=1", "GIT_TRACE=1")
	cmd.Stdin = os.Stdin

	out, err := cmd.Output()
	if err != nil {
		return err
	}

	var headHash, headTarget string
	refs := make(map[string]string)
	for _, line := range strings.Split(string(out), "\n") {
		log.Printf("line: %q\n", line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)

		if fields[1] == "HEAD" {
			headHash = fields[0]
			continue
		}

		if headHash != "" && fields[0] == headHash {
			headTarget = fields[1]
		}

		refs[fields[1]] = fields[0]
	}

	for objectname, refname := range refs {
		log.Printf("%s %s\n", refname, objectname)
		_, _ = os.Stdout.WriteString(fmt.Sprintf("%s %s\n", refname, objectname))
	}

	if headTarget != "" {
		log.Printf("@%s HEAD \n", headTarget)
		_, _ = os.Stdout.WriteString(fmt.Sprintf("@%s HEAD \n", headTarget))
	}

	_, _ = os.Stdout.WriteString("\n")
	return nil
}

func handleFetch(remoteUrl, args string) error {
	cmdParts := append([]string{"fetch", remoteUrl}, strings.Fields(args)...)
	log.Printf("git fetch %s %s", remoteUrl, args)
	cmd := exec.Command("git", cmdParts...)
	cmd.Env = append(os.Environ(), "GIT_TRACE_PACKET=1", "GIT_TRACE=1")
	cmd.Stdin = os.Stdin

	out, err := cmd.Output()
	log.Printf("output: %s\n", out)
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(out)
	return err
}

func handlePush(remoteUrl string) error {
	log.Printf("git push %s", remoteUrl)
	cmd := exec.Command("git", "push", remoteUrl)
	cmd.Env = append(os.Environ(), "GIT_TRACE_PACKET=1", "GIT_TRACE=1")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func doHTTP(remoteName, remoteUrl, scheme string) error {
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
