package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

func NewLog() *os.File {
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

func Main() (er error) {

	if len(os.Args) < 3 {
		return fmt.Errorf("Usage: git-remote-nostr remote-name url")
	}

	remoteName := os.Args[1]
	url := os.Args[2]
	log.Printf("running git-remote-nostr: %s:%s\n", remoteName, url)
	// Add "path" to the import list

	localdir := path.Join(os.Getenv("GIT_DIR"), "nostr", remoteName)

	if err := os.MkdirAll(localdir, 0755); err != nil {
		return err
	}

	refspec := fmt.Sprintf("refs/heads/*:refs/nostr/%s/*", remoteName)

	if err := os.Setenv("GIT_DIR", path.Join(url, ".git")); err != nil {
		return err
	}

	gitmarks := path.Join(localdir, "git.marks")
	nostrmarks := path.Join(localdir, "nostr.marks")

	if err := Touch(gitmarks); err != nil {
		return err
	}

	if err := Touch(nostrmarks); err != nil {
		return err
	}

	// add "io/ioutil" to imports

	originalGitmarks, err := os.ReadFile(gitmarks)
	if err != nil {
		return err
	}

	originalNostrmarks, err := os.ReadFile(nostrmarks)
	if err != nil {
		return err
	}

	defer func() {
		if er != nil {
			_ = os.WriteFile(gitmarks, originalGitmarks, 0666)
			_ = os.WriteFile(nostrmarks, originalNostrmarks, 0666)
		}
	}()

	// Add "bufio" to import list.

	stdinReader := bufio.NewReader(os.Stdin)

	for {
		// Note that command will include the trailing newline.
		command, err := stdinReader.ReadString('\n')
		if err != nil {
			return err
		}
		log.Printf("command: %q\n", command)

		switch {
		case command == "capabilities\n":
			fmt.Printf("import\n")
			fmt.Printf("export\n")
			fmt.Printf("refspec %s\n", refspec)
			fmt.Printf("*import-marks %s\n", gitmarks)
			fmt.Printf("*export-marks %s\n", gitmarks)
			fmt.Printf("\n")
		case command == "list\n":
			refs, err := GitListRefs()
			if err != nil {
				log.Printf("command list: GetListRefs: %v\n", err)
				return fmt.Errorf("command list: %v", err)
			}

			head, err := GitSymbolicRef("HEAD")
			if err != nil {
				log.Printf("command list: GitSymbolicRef: %v\n", err)
				return fmt.Errorf("command list: %v", err)
			}

			for refname := range refs {
				log.Printf("? %s\n", refname)
				fmt.Printf("? %s\n", refname)
			}

			log.Printf("@%s HEAD\n", head)
			fmt.Printf("@%s HEAD\n", head)
			log.Printf("\n")
			fmt.Printf("\n")
		case strings.HasPrefix(command, "import "):
			refs := make([]string, 0)

			for {
				// Have to make sure to trim the trailing newline.
				ref := strings.TrimSpace(strings.TrimPrefix(command, "import "))

				refs = append(refs, ref)
				command, err = stdinReader.ReadString('\n')
				if err != nil {
					return err
				}

				if !strings.HasPrefix(command, "import ") {
					break
				}
			}

			fmt.Printf("feature import-marks=%s\n", gitmarks)
			fmt.Printf("feature export-marks=%s\n", gitmarks)
			fmt.Printf("feature done\n")

			args := []string{
				"fast-export",
				"--import-marks", nostrmarks,
				"--export-marks", nostrmarks,
				"--refspec", refspec}
			args = append(args, refs...)

			cmd := exec.Command("git", args...)
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout

			if err := cmd.Run(); err != nil {
				return fmt.Errorf("command import: git fast-export: %v", err)
			}

			fmt.Printf("done\n")
		case command == "export\n":
			beforeRefs, err := GitListRefs()
			if err != nil {
				return fmt.Errorf("command export: collecting before refs: %v", err)
			}

			cmd := exec.Command("git", "fast-import", "--quiet",
				"--import-marks="+nostrmarks,
				"--export-marks="+nostrmarks)

			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin

			if err := cmd.Run(); err != nil {
				return fmt.Errorf("command export: git fast-import: %v", err)
			}

			afterRefs, err := GitListRefs()
			if err != nil {
				return fmt.Errorf("command export: collecting after refs: %v", err)
			}

			for refname, objectname := range afterRefs {
				if beforeRefs[refname] != objectname {
					fmt.Printf("ok %s\n", refname)
				}
			}

			fmt.Printf("\n")
		case command == "\n":
			return nil
		default:
			return fmt.Errorf("Received unknown command %q", command)
		}
	}
}

// Add "os/exec" and "bytes" to the import list.

// GitListRefs Returns a map of refnames to objectnames.
func GitListRefs() (map[string]string, error) {
	out, err := exec.Command(
		"git", "for-each-ref", "--format=%(objectname) %(refname)",
		"refs/heads/",
	).Output()
	log.Printf("GitListRefs: %s\n", out)
	if err != nil {
		return nil, err
	}

	lines := bytes.Split(out, []byte{'\n'})
	refs := make(map[string]string, len(lines))

	for _, line := range lines {
		fields := bytes.Split(line, []byte{' '})

		if len(fields) < 2 {
			break
		}

		refs[string(fields[1])] = string(fields[0])
	}

	return refs, nil
}

func GitSymbolicRef(name string) (string, error) {
	out, err := exec.Command("git", "symbolic-ref", name).Output()
	log.Printf("GitSymbolicRef: %s\n", out)
	if err != nil {
		return "", fmt.Errorf("GitSymbolicRef: git symbolic-ref %s: %v", name, err)
	}

	return string(bytes.TrimSpace(out)), nil
}

// Touch Create path as an empty file if it doesn't exist, otherwise do nothing.
// This works by opening a file in exclusive mode; if it already exists,
// an error will be returned rather than truncating it.
func Touch(path string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
	if os.IsExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	return file.Close()
}

func main() {
	logOut := NewLog()
	defer logOut.Close()

	if err := Main(); err != nil {
		log.Fatal(err)
	}

	log.Printf("\n\n===============================================================================\n\n")
}
