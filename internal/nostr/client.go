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

type Client struct {
	pvtKey string
	pubKey string
}

var _ IClient = &Client{}

func NewClient(pvtKey, pubKey string) *Client {
	return &Client{
		pvtKey: pvtKey,
		pubKey: pubKey,
	}
}

func (c *Client) PvtKey() string {
	return c.pvtKey
}

func (c *Client) PubKey() string {
	return c.pubKey
}

func (c *Client) ResolveWithNip05(remote *domain.Remote) (*domain.Remote, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, name, err := nip05.Fetch(ctx, remote.Nip05())
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("failed to fetch nip05: %s", remote.Nip05()))
	}

	authorPubKey, ok := resp.Names[name]
	if !ok {
		return nil, errors.New("no public key found")
	}

	relays := append([]string{remote.PrimaryRelay()}, resp.Relays[authorPubKey]...)

	return c.ResolveWithFilters(relays, []nostr.Filter{{
		Kinds:   []int{nostr.KindRepositoryAnnouncement},
		Authors: []string{authorPubKey},
		Tags: nostr.TagMap{
			"d": {remote.Path()},
		},
	}})
}

func (c *Client) ResolveWithFilters(relays []string, filters nostr.Filters) (*domain.Remote, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	pool := nostr.NewSimplePool(ctx)
	for ev := range pool.SubManyEoseNonUnique(ctx, relays, filters) {
		remote, err := domain.ParseRemoteFromEvent(ev.Event)
		if err != nil {
			continue
		}

		return remote, nil
	}

	return nil, errors.New("no remote found")
}

func (c *Client) Auth(remote *domain.Remote, key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	primary := remote.PrimaryRelay()
	relay, err := nostr.RelayConnect(ctx, primary)
	if err != nil {
		return errors.Join(err, fmt.Errorf("failed to connect to relay: %s", primary))
	}
	defer relay.Close()

	return relay.Auth(ctx, func(evt *nostr.Event) error {
		evt.Content = key
		evt.Tags = append(evt.Tags, nostr.Tags{{"path", remote.Path()}}...)
		return evt.Sign(c.pvtKey)
	})
}
