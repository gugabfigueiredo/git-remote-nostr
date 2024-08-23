package domain

import (
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/nbd-wtf/go-nostr"
	"strings"
)

type RemoteService interface {
	ResolveRemote(remoteRaw string) (*Remote, error)
	Auth(remote *Remote, key string) error
}

type Remote struct {
	*transport.Endpoint
	Event *nostr.Event
}

func ParseRemote(address string) (*Remote, error) {
	e, err := transport.NewEndpoint(address)
	if err != nil {
		return nil, fmt.Errorf("failed to parse remote: %s", address)
	}

	return &Remote{Endpoint: e}, nil
}

func ParseRemoteFromEvent(event *nostr.Event) (*Remote, error) {
	if evOk, err := event.CheckSignature(); !evOk || err != nil {
		return nil, errors.Join(err, errors.New("event signature is invalid"))
	}

	remoteTag := event.Tags.GetFirst([]string{"remote"})
	if remoteTag == nil {
		// supporting current nip-34 standards
		remoteTag = event.Tags.GetFirst([]string{"clone"})
		if remoteTag == nil {
			return nil, errors.New("no remote tag found")
		}
	}

	remote, err := ParseRemote(remoteTag.Value())
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to parse remote"))
	}

	remote.Event = event

	return remote, nil
}

func (r *Remote) Nip05() string {
	if r.User == "" {
		return "_@" + r.Host
	}
	return r.User + "@" + r.Host
}

func (r *Remote) Nip19() string {
	if r.User != "" {
		return r.User
	}

	return r.Host
}

func (r *Remote) PrimaryRelay() string {
	hostAndPort := r.Host
	port := r.Port
	if port > 0 && port != 22 {
		hostAndPort = fmt.Sprintf("%s:%d", hostAndPort, port)
	}
	return "wss://" + hostAndPort
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
