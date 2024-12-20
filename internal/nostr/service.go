package nostr

import (
	"errors"
	"fmt"
	"github.com/gugabfigueiredo/git-remote-nostr/internal/domain"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip05"
	"github.com/nbd-wtf/go-nostr/nip19"
)

type IClient interface {
	Configer
	Resolver
	Auther
}

type Configer interface {
	PvtKey() string
	PubKey() string
}

type Resolver interface {
	ResolveWithNip05(remote *domain.Remote) (*domain.Remote, error)
	ResolveWithFilters(relays []string, filters nostr.Filters) (*domain.Remote, error)
}

type Auther interface {
	Auth(remote *domain.Remote) error
}

type Service struct {
	IClient
}

func NewService(client IClient) *Service {
	return &Service{client}
}

func (s *Service) ResolveRemote(remoteRaw string) (*domain.Remote, error) {
	remote, err := domain.ParseRemote(remoteRaw)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("failed to parse remote: %s", remoteRaw))
	}

	if prefix, v, err := nip19.Decode(remote.Nip19()); err == nil {
		switch prefix {
		case "npub":
			return s.ResolveWithFilters([]string{remote.PrimaryRelay()}, []nostr.Filter{{
				Kinds:   []int{nostr.KindRepositoryAnnouncement},
				Authors: []string{v.(string)},
				Tags: nostr.TagMap{
					"d": {remote.Path()},
				},
			}})
		case "nevent":
			nevent := v.(nostr.EventPointer)
			return s.ResolveWithFilters(nevent.Relays, []nostr.Filter{{
				Kinds: []int{nostr.KindRepositoryAnnouncement},
				IDs:   []string{nevent.ID},
			}})
		case "nprofile":
			nprofile := v.(nostr.ProfilePointer)
			return s.ResolveWithFilters(nprofile.Relays, []nostr.Filter{{
				Authors: []string{nprofile.PublicKey},
				Kinds:   []int{nostr.KindRepositoryAnnouncement},
				Tags: nostr.TagMap{
					"d": {remote.Path()},
				},
			}})
		case "note":
			return s.ResolveWithFilters([]string{remote.PrimaryRelay()}, nostr.Filters{{
				Kinds: []int{nostr.KindRepositoryAnnouncement},
				IDs:   []string{v.(string)},
			}})
		case "naddr":
			entity := v.(nostr.EntityPointer)
			return s.ResolveWithFilters([]string{remote.PrimaryRelay()}, nostr.Filters{{
				Kinds:   []int{nostr.KindRepositoryAnnouncement},
				Authors: []string{entity.PublicKey},
				Tags: nostr.TagMap{
					"d": {entity.Identifier},
				},
			}})
		default:
			return nil, fmt.Errorf("unsupported nip-19 prefix: %s", remote.String())
		}
	}

	if nostr.IsValidPublicKey(remote.User) {
		return s.ResolveWithFilters([]string{remote.PrimaryRelay()}, []nostr.Filter{{
			Kinds:   []int{nostr.KindRepositoryAnnouncement},
			Authors: []string{remote.User},
			Tags: nostr.TagMap{
				"d": {remote.Path()},
			},
		}})
	}

	if nip05.IsValidIdentifier(remote.Nip05()) {
		return s.ResolveWithNip05(remote)
	}

	return nil, fmt.Errorf("unsupported nostr remote url: %s", remote.String())
}
