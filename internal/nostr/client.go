package nostr

import (
	"context"
	"errors"
	"fmt"
	"github.com/gugabfigueiredo/git-remote-nostr/internal/domain"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip05"
	"time"
)

type IClient interface {
	resolveWithNip05(remote *domain.Remote) (*domain.Remote, error)
	resolveWithFilters(relays []string, filters nostr.Filters) (*domain.Remote, error)
}

type Client struct {
}

var _ IClient = &Client{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) resolveWithNip05(remote *domain.Remote) (*domain.Remote, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, name, err := nip05.Fetch(ctx, remote.Nip05())
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("failed to fetch nip05: %s", remote.Nip05()))
	}

	pubKey, ok := resp.Names[name]
	if !ok {
		return nil, errors.New("no public key found")
	}

	relays := append([]string{remote.PrimaryRelay()}, resp.Relays[pubKey]...)

	return c.resolveWithFilters(relays, []nostr.Filter{{
		Kinds:   []int{nostr.KindRepositoryAnnouncement},
		Authors: []string{pubKey},
		Tags: map[string][]string{
			"d": {remote.Path()},
		},
	}})
}

func (c *Client) resolveWithFilters(relays []string, filters nostr.Filters) (*domain.Remote, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	for _, relay := range relays {
		if !nostr.IsValidRelayURL(relay) {
			continue
		}

		conn, err := nostr.RelayConnect(ctx, relay)
		if err != nil {
			continue
		}

		// create a subscription and submit to relay
		// results will be returned on the sub.Events channel
		sub, _ := conn.Subscribe(ctx, filters)
		for {
			select {
			case <-sub.EndOfStoredEvents:
				// Handle end of stored events here if needed
				continue
			case ev, ok := <-sub.Events:
				if !ok {
					// The Events channel is closed, so exit the loop
					continue
				}

				if remoteTag := ev.Tags.GetFirst([]string{"clone"}); remoteTag != nil {
					return domain.ParseRemote(remoteTag.Value())
				}

				// Handle the event here
				if remoteTag := ev.Tags.GetFirst([]string{"remote"}); remoteTag != nil {
					return domain.ParseRemote(remoteTag.Value())
				}
			}
		}
		sub.Unsub()
	}

	return nil, errors.New("no remote found")
}
