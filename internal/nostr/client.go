package nostr

import (
	"context"
	"errors"
	"fmt"
	"github.com/gugabfigueiredo/git-remote-nostr/internal/domain"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip05"
	"github.com/nbd-wtf/go-nostr/nip11"
	"slices"
	"strings"
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

func (c *Client) Auth(remote *domain.Remote) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	primary := remote.PrimaryRelay()
	info, err := nip11.Fetch(ctx, primary)
	if err != nil {
		// if we can't fetch nip11, we must assume it's not relay; skip nostr auth
		return nil
	}
	if !info.Limitation.AuthRequired {
		// no way to auth with this relay
		return nil
	}
	if !slices.Contains(info.SupportedNIPs, 34) {
		// not a nostr-git relay we cannot complete auth
		return errors.New("remote is not a nostr-git relay")
	}

	relay, err := nostr.RelayConnect(ctx, primary)
	if err != nil {
		return errors.Join(err, fmt.Errorf("failed to connect to relay: %s", primary))
	}
	defer relay.Close()

	authMethod := resolveAuthMethod(info, remote)

	return relay.Auth(ctx, func(evt *nostr.Event) error {
		if err = authMethod(evt); err != nil {
			return err
		}
		return evt.Sign(c.pvtKey)
	})
}

func resolveAuthMethod(info nip11.RelayInformationDocument, remote *domain.Remote) func(evt *nostr.Event) error {
	var method string
	for _, tag := range info.Tags {
		if strings.HasPrefix(tag, "auth-method:") {
			method = strings.TrimPrefix(tag, "auth-method:")
			break
		}
	}
	if tag := remote.Event.Tags.GetFirst([]string{"auth-method"}); tag != nil {
		method = tag.Value()
	}
	switch method {
	case "ssh-pubkey":
		return func(evt *nostr.Event) error {
			// generate ssh-sso key pair
			// store the private key locally and add to key agent
			// sign the event with the public key and path
			evt.Content = "key"
			evt.Tags = append(evt.Tags, nostr.Tags{{"path", remote.Path()}, {"remote-id", remote.Event.ID}}...)
			return nil
		}
	default:
		return func(evt *nostr.Event) error {
			return nil
		}
	}
}
