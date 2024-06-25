package domain

import (
	"fmt"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"strings"
)

type RemoteService interface {
	ResolveRemote(remoteRaw string) (*Remote, error)
}

type Remote struct {
	*transport.Endpoint
}

func ParseRemote(address string) (*Remote, error) {
	e, err := transport.NewEndpoint(address)
	if err != nil {
		return nil, fmt.Errorf("failed to parse remote: %s", address)
	}

	return &Remote{Endpoint: e}, nil
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
