package nostr

import (
	"errors"
	"fmt"
	"github.com/gugabfigueiredo/git-remote-nostr/internal/domain"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip05"
	"github.com/nbd-wtf/go-nostr/nip19"
)

type Service struct {
	IClient
}

func NewService(client IClient) *Service {
	return &Service{client}
}

func (n *Service) ResolveRemote(remoteRaw string) (*domain.Remote, error) {
	remote, err := domain.ParseRemote(remoteRaw)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("failed to parse remote: %s", remoteRaw))
	}

	if prefix, v, err := nip19.Decode(remote.Nip19()); err == nil {
		switch prefix {
		case "npub":
			return n.resolveWithFilters([]string{remote.PrimaryRelay()}, []nostr.Filter{{
				Kinds:   []int{nostr.KindRepositoryAnnouncement},
				Authors: []string{v.(string)},
				Tags: map[string][]string{
					"d": {remote.Path()},
				},
			}})
		case "nevent":
			nevent := v.(nostr.EventPointer)
			return n.resolveWithFilters(nevent.Relays, []nostr.Filter{{
				Kinds: []int{nostr.KindRepositoryAnnouncement},
				IDs:   []string{nevent.ID},
			}})
		case "nprofile":
			nprofile := v.(nostr.ProfilePointer)
			return n.resolveWithFilters(nprofile.Relays, []nostr.Filter{{
				Authors: []string{nprofile.PublicKey},
				Kinds:   []int{nostr.KindRepositoryAnnouncement},
				Tags: map[string][]string{
					"d": {remote.Path()},
				},
			}})
		case "note":
			return n.resolveWithFilters([]string{remote.PrimaryRelay()}, nostr.Filters{{
				Kinds: []int{nostr.KindRepositoryAnnouncement},
				IDs:   []string{v.(string)},
			}})
		case "naddr":
			entity := v.(nostr.EntityPointer)
			return n.resolveWithFilters([]string{remote.PrimaryRelay()}, nostr.Filters{{
				Kinds:   []int{nostr.KindRepositoryAnnouncement},
				Authors: []string{entity.PublicKey},
				Tags: map[string][]string{
					"d": {entity.Identifier},
				},
			}})
		default:
			return nil, fmt.Errorf("unsupported nip-19 prefix: %s", remote.String())
		}
	}

	if nostr.IsValidPublicKey(remote.User) {
		return n.resolveWithFilters([]string{remote.PrimaryRelay()}, []nostr.Filter{{
			Kinds:   []int{nostr.KindRepositoryAnnouncement},
			Authors: []string{remote.User},
			Tags: map[string][]string{
				"d": {remote.Path()},
			},
		}})
	}

	if nip05.IsValidIdentifier(remote.User + "@" + remote.Host) {
		return n.resolveWithNip05(remote)
	}

	return nil, fmt.Errorf("unsupported nostr remote url: %s", remote.String())
}
