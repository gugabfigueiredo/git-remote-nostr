package main

import (
	"bufio"
	"fmt"
	"github.com/melbahja/goph"
	"log"
	"os"
	"os/exec"
	"strings"
)

var client *goph.Client

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
			fmt.Println("connect")
			fmt.Println("fetch")
			fmt.Println("push")
			fmt.Println()
		case strings.HasPrefix(command, "fetch "):
			log.Printf("fetch command: %q\n", command)
		case strings.HasPrefix(command, "push "):
			log.Printf("push command: %q\n", command)
		case strings.HasPrefix(command, "connect "):
			if err := handleSSHConnect(command, remoteUrl); err != nil {
				return err
			}
			fmt.Println()
		case command == "list\n":
			if err := handleSSHList(remoteUrl); err != nil {
				return err
			}
			fmt.Println()
		case command == "\n":
			fmt.Println()
		default:
			return fmt.Errorf("unknown command: %q", command)
		}
	}
}

func handleSSHConnect(command, remoteUrl string) error {
	var head, tail string
	lineFields(command, &head, &command, &tail)

	switch command {
	case "git-receive-pack":
		log.Printf("ssh git receive-pack '%s %s'", remoteUrl, tail)
		fmt.Println()
	case "git-send-pack":
		log.Printf("ssh git-send-pack '%s %s'", remoteUrl, tail)
		fmt.Println()
	case "git-upload-pack":
		log.Printf("ssh git-upload-pack '%s %s'", remoteUrl, tail)
		//err := doUploadPack()
		err := doFetchPack(remoteUrl, tail)
		if err != nil {
			return err
		}
		fmt.Println()
	case "git-fetch-pack":
		log.Printf("ssh git-fetch-pack '%s %s'", remoteUrl, tail)
		fmt.Println()
	default:
		return fmt.Errorf("unknown connect command: %q", strings.Join([]string{command, tail}, " "))
	}
	return nil
}

//func doUploadPack() error {
//	log.Printf("git-upload-pack")
//	auth, err := goph.UseAgent()
//	if err != nil {
//		panic(err)
//	}
//
//	// Do something with auth
//	if client == nil {
//		client, err = goph.New("git", "github.com", auth)
//		client.RemoteAddr()
//		if err != nil {
//			panic(err)
//		}
//	}
//
//	// Do something with client
//	cmd, err := client.Command("git-upload-pack", "'gugabfigueiredo/git-remote-nostr.git'")
//	if err != nil {
//		panic(err)
//	}
//
//	// Do something with cmd
//	out, err := cmd.Output()
//	if err != nil {
//		return err
//	}
//
//	lines := strings.Split(string(out), "\n")
//	for i, line := range lines {
//		line = line[4:]
//
//		if i == 0 {
//			line, capabilites := getSSHCapabilities(line)
//		}
//
//		if line != "" {
//			log.Printf("line: %q\n", line)
//			_, _ = os.Stdout.WriteString(fmt.Sprintf("%s\n", line))
//		}
//	}
//	return err
//}

func getSSHCapabilities(line string) (string, []string) {
	split := strings.Split(line, "\x00")
	line = split[0]
	capabilities := strings.Split(split[1], " ")
	return line, capabilities
}

//func parseSSHUrl(url string) (login, path string, err error) {
//	// Regular expression to match both SCP-like and SSH URI formats
//	var re = regexp.MustCompile(`^(?:(?:(\w+)@?[\w\.]+):([^@]+)$)|(?:ssh:\/\/(\w+)@([\w\.]+)\/([^@]+)$)`)
//	matches := re.FindStringSubmatch(url)
//
//	if matches == nil {
//		return "", "", fmt.Errorf("invalid SSH URL format")
//	}
//
//	// Check which format was matched
//	if matches[1] != "" || matches[2] != "" {
//		// SCP-like syntax
//		user = matches[1]
//		host = matches[2]
//		path = matches[3]
//	} else {
//		// SSH URI syntax
//		user = matches[4]
//		host = matches[5]
//		path = matches[6]
//	}
//
//	return user, host, path, nil
//}

func handleSSHList(remoteUrl string) error {
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

func handleSSHFetch(command, remoteUrl string) error {
	var args string
	lineFields(command, &command, &args)
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

func handleSSHPush(remoteUrl string) error {
	log.Printf("git push %s", remoteUrl)
	cmd := exec.Command("git", "push", remoteUrl)
	cmd.Env = append(os.Environ(), "GIT_TRACE_PACKET=1", "GIT_TRACE=1")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
