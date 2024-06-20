package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/go-git/go-git/v5/plumbing/transport"
	goNostr "github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip05"
	"github.com/pkg/errors"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

//var env struct {
//	Nostr nostr.Config
//}

//func init() {
//	envconfig.MustProcess("GIT_NOSTR", &env.Nostr)
//}

func main() {
	args := os.Args
	if len(args) < 3 {
		log.Fatal("Usage: git-remote-nostr <remoteName> <remoteUrl>")
	}

	remote, err := resolveRemote("nostr://827742da08c7911862c23ddd4758a57ad4c3bb7c9b89034df0485674786f1644@localhost:3334/git-remote-nostr")
	if err != nil {
		log.Fatalf("Error parsing remote: %v", err)
	}

	switch remote.Protocol {
	case "ssh", "nostr":
		fmt.Println("remote is ssh or nostr", remote.String())
		if err = doSSH(remote); err != nil {
			log.Fatalf("failed to setup remote transport: %v", err)
		}
	default:
		if err = doDefault(args[1], remote); err != nil {
			log.Fatalf("failed to setup remote transport: %v", err)
		}
	}
}

func doDefault(remoteName string, remote *Remote) error {
	cmd := exec.Command("git", "remote-"+remote.Protocol, remoteName, remote.String())
	return doCMD(cmd)
}

func doCMD(cmd *exec.Cmd) error {
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Println()
	return cmd.Run()
}

func doSSH(remote *Remote) error {
	stdinReader := bufio.NewReader(os.Stdin)
	for {
		command, err := stdinReader.ReadString('\n')
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return errors.Wrap(err, "failed to read command")
		}

		switch {
		case command == "capabilities\n":
			fmt.Println("connect")
			fmt.Println()
		case command == "connect git-upload-pack\n":
			//if err := doConnect("git-upload-pack", remote); err != nil {
			//	return err
			//}
			fmt.Println("running git-upload-pack for", remote.String())
		case command == "connect git-receive-pack\n":
			//if err := doConnect("git-receive-pack", remote); err != nil {
			//	return err
			//}
			fmt.Println("running git-receive-pack for", remote.String())
		default:
			return errors.Errorf("unknown command: %s", command)
		}
	}
}

func doConnect(command string, remote *Remote) error {
	cmd := exec.Command("ssh", remote.Login(), command, remote.Path())
	return doCMD(cmd)
}

func resolveRemote(remoteRaw string) (*Remote, error) {
	// if the remote is git friendly, we can use it directly
	remote, err := parseRemote(remoteRaw)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse remote")
		//return resolveNostr(remoteRaw)
	}

	if remote.Protocol == "nostr" {
		return fetchNostrRemote(remote)
	}

	return nil, errors.New("remote is invalid")
}

func resolveNostr(remoteRaw string) (*Remote, error) {
	return nil, nil
}

func fetchNostrRemote(remote *Remote) (*Remote, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if nip05.IsValidIdentifier(remote.User + "@" + remote.Host) {
		return fetchFromNip05(ctx, remote)
	}

	if goNostr.IsValidPublicKey(remote.User) {
		return fetchWithPubKey(ctx, remote.User, remote.PrimaryRelay(), remote.Path())
	}

	//if goNostr.IsValid32ByteHex(remote.User) {
	//	return fetchWithPubKey(ctx, makeNpub)
	//}

	return nil, errors.New("invalid nostr remote")
}

type Remote struct {
	*transport.Endpoint
}

func parseRemote(address string) (*Remote, error) {
	e, err := transport.NewEndpoint(address)
	if err != nil {
		return nil, err
	}

	return &Remote{Endpoint: e}, nil
}

func (r *Remote) Nip05() string {
	if r.User == "" {
		return "_@" + r.Host
	}
	return r.User + "@" + r.Host
}

func (r *Remote) PrimaryRelay() string {
	hostAndPort := r.Host
	port := r.Port
	if port != 22 {
		hostAndPort = fmt.Sprintf("%s:%d", hostAndPort, port)
	}
	return "ws://" + hostAndPort
}

func (r *Remote) Login() string {
	hostAndPort := r.Host
	port := r.Port
	if port != 22 {
		hostAndPort = fmt.Sprintf("%s:%d", hostAndPort, port)
	}
	return r.User + "@" + hostAndPort
}

func (r *Remote) Path() string {
	return strings.TrimPrefix(r.Endpoint.Path, "/")
}

func fetchFromNip05(ctx context.Context, nostrRemote *Remote) (*Remote, error) {
	resp, name, err := nip05.Fetch(ctx, nostrRemote.Nip05())
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch from nip05")
	}

	pubKey, ok := resp.Names[name]
	if !ok {
		return nil, errors.New("no public key found")
	}

	relays := append([]string{nostrRemote.PrimaryRelay()}, resp.Relays[pubKey]...)

	for _, relay := range relays {
		if !goNostr.IsValidRelayURL(relay) {
			continue
		}

		remote, err := fetchWithPubKey(ctx, pubKey, relay, nostrRemote.Path())
		if err != nil {
			continue
		}

		return remote, nil
	}

	return nil, errors.New("no remote found")
}

func fetchWithPubKey(ctx context.Context, pubKey, relay, repoId string) (*Remote, error) {
	// now we need to use pubKey to search for the remote
	conn, err := goNostr.RelayConnect(ctx, relay)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to relay")
	}

	filters := []goNostr.Filter{{
		Authors: []string{pubKey},
		Tags: map[string][]string{
			"d": {repoId},
		},
	}}

	// create a subscription and submit to relay
	// results will be returned on the sub.Events channel
	sub, _ := conn.Subscribe(ctx, filters)
	defer sub.Unsub()
	for {
		select {
		case <-sub.EndOfStoredEvents:
			// Handle end of stored events here if needed
			return nil, errors.New("no remote found")
		case ev, ok := <-sub.Events:
			if !ok {
				// The Events channel is closed, so exit the loop
				return nil, errors.New("no remote found")
			}

			// Handle the event here
			if remoteTag := ev.Tags.GetFirst([]string{"remote"}); remoteTag != nil {
				return parseRemote(remoteTag.Value())
			}
		}
	}
}
