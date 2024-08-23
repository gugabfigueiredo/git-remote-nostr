package nostr

import (
	"fmt"
	"github.com/gugabfigueiredo/git-remote-nostr/internal/domain"
	"github.com/nbd-wtf/go-nostr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestService_ResolveRemote(t *testing.T) {

	tt := []struct {
		name               string
		remoteRaw          string
		client             *ClientStub
		expectedCalledWith any
		expectedRemote     string
		expectedError      error
	}{
		{
			name:      "nostr: pubkey@relay/path",
			remoteRaw: "nostr://96beaaaeffcf4ca1b1e988bc5f112602b20945b67dc27339af31bd0854bcbf47@relay/path",
			client: &ClientStub{
				resolveWithFiltersResponse: "git@github.str:username/repo",
			},
			expectedCalledWith: struct {
				relays  []string
				filters nostr.Filters
			}{
				relays: []string{"ws://relay"},
				filters: nostr.Filters{{
					IDs:     []string(nil),
					Kinds:   []int{30617},
					Authors: []string{"96beaaaeffcf4ca1b1e988bc5f112602b20945b67dc27339af31bd0854bcbf47"},
					Tags:    nostr.TagMap{"d": []string{"path"}},
				}},
			},
			expectedRemote: "ssh://git@github.str/username/repo",
		},
		{
			name:      "nostr: pubkey@relay:port/path",
			remoteRaw: "nostr://96beaaaeffcf4ca1b1e988bc5f112602b20945b67dc27339af31bd0854bcbf47@relay:2222/path",
			client: &ClientStub{
				resolveWithFiltersResponse: "git@github.str:username/repo",
			},
			expectedCalledWith: struct {
				relays  []string
				filters nostr.Filters
			}{
				relays: []string{"ws://relay:2222"},
				filters: nostr.Filters{{
					Kinds:   []int{30617},
					Authors: []string{"96beaaaeffcf4ca1b1e988bc5f112602b20945b67dc27339af31bd0854bcbf47"},
					Tags:    nostr.TagMap{"d": []string{"path"}},
				}},
			},
			expectedRemote: "ssh://git@github.str/username/repo",
		},
		{
			name:      "nostr: nip05@relay/path",
			remoteRaw: "nostr://nip05@relay.str/path",
			client: &ClientStub{
				resolveWithNip05Response: "git@github.str:username/repo",
			},
			expectedCalledWith: "nostr://nip05@relay.str/path",
			expectedRemote:     "ssh://git@github.str/username/repo",
		},
		{
			name:      "nostr: nip19npub@relay/path",
			remoteRaw: "nostr://npub10h2y7sgrwfrjscvl54ghj8grgl2h0qhqz4d5s3utejsg3xpnhansrk7tvg@relay/path",
			client: &ClientStub{
				resolveWithFiltersResponse: "git@github.str:username/repo",
			},
			expectedCalledWith: struct {
				relays  []string
				filters nostr.Filters
			}{
				relays: []string{"ws://relay"},
				filters: nostr.Filters{{
					Kinds:   []int{30617},
					Authors: []string{"7dd44f4103724728619fa551791d0347d57782e0155b48478bcca0889833bf67"},
					Tags:    nostr.TagMap{"d": []string{"path"}},
				}},
			},
			expectedRemote: "ssh://git@github.str/username/repo",
		},
		{
			name:      "nostr: nip19nevent",
			remoteRaw: "nostr://nevent1qqs9eksmfqsrgsylyxyrrggav5hddxlapq66zzqx0n5r2v6pj95dzvcppemhxw309aex2mrp0yh8xarjqgsteflh5aqehn7zlh2h9khcskd6glsppfklfaqj656gpd75p0hvkkqqdkh36",
			client: &ClientStub{
				resolveWithFiltersResponse: "git@github.str:username/repo",
			},
			expectedCalledWith: struct {
				relays  []string
				filters nostr.Filters
			}{
				relays: []string{"ws://relay.str"},
				filters: nostr.Filters{{
					IDs:   []string{"5cda1b482034409f218831a11d652ed69bfd0835a108067ce83533419168d133"},
					Kinds: []int{30617},
				}},
			},
			expectedRemote: "ssh://git@github.str/username/repo",
		},
		{
			name:      "nostr: nip19nprofile/path",
			remoteRaw: "nostr://nprofile1qqsqjn3ns9dfrhgjm9tffgjsz5wz8znjz5alm0uhjlxc32ysfp50xtsppfmhxw309aex2mrp0y3tse0l/path",
			client: &ClientStub{
				resolveWithFiltersResponse: "git@github.str:username/repo",
			},
			expectedCalledWith: struct {
				relays  []string
				filters nostr.Filters
			}{
				relays: []string{"ws://relay"},
				filters: nostr.Filters{{
					Kinds:   []int{30617},
					Authors: []string{"094e33815a91dd12d95694a250151c238a72153bfdbf9797cd88a8904868f32e"},
					Tags:    nostr.TagMap{"d": []string{"path"}},
				}},
			},
			expectedRemote: "ssh://git@github.str/username/repo",
		},
		{
			name:      "pubkey@relay:path",
			remoteRaw: "96beaaaeffcf4ca1b1e988bc5f112602b20945b67dc27339af31bd0854bcbf47@relay:path",
			client: &ClientStub{
				resolveWithFiltersResponse: "git@github.str:username/repo",
			},
			expectedCalledWith: struct {
				relays  []string
				filters nostr.Filters
			}{
				relays: []string{"ws://relay"}, filters: nostr.Filters{{
					Kinds:   []int{30617},
					Authors: []string{"96beaaaeffcf4ca1b1e988bc5f112602b20945b67dc27339af31bd0854bcbf47"},
					Tags:    nostr.TagMap{"d": []string{"path"}},
				}},
			},
			expectedRemote: "ssh://git@github.str/username/repo",
		},
		{
			name:      "nip05@relay:path",
			remoteRaw: "nip05@relay.str:path",
			client: &ClientStub{
				resolveWithNip05Response: "git@github.str:username/repo",
			},
			expectedCalledWith: "ssh://nip05@relay.str/path",
			expectedRemote:     "ssh://git@github.str/username/repo",
		},
		{
			name:      "nip19npub@relay:path",
			remoteRaw: "npub1xvnhj4mrsxpcqktp72cqrstv55e2m0w8m2ypap8g3quy8ljhhefq505e77@relay:path",
			client: &ClientStub{
				resolveWithFiltersResponse: "git@github.str:username/repo",
			},
			expectedCalledWith: struct {
				relays  []string
				filters nostr.Filters
			}{
				relays: []string{"ws://relay"},
				filters: nostr.Filters{{
					Kinds:   []int{30617},
					Authors: []string{"33277957638183805961f2b001c16ca532adbdc7da881e84e8883843fe57be52"},
					Tags:    nostr.TagMap{"d": []string{"path"}},
				}},
			},
			expectedRemote: "ssh://git@github.str/username/repo",
		},
		{
			name:          "invalid nostr remote",
			remoteRaw:     "nostr://google.com/not/a/valid/path",
			client:        &ClientStub{},
			expectedError: fmt.Errorf("unsupported nostr remote url: nostr://google.com/not/a/valid/path"),
		},
		{
			name:      "fail to resolve with filers",
			remoteRaw: "nostr://96beaaaeffcf4ca1b1e988bc5f112602b20945b67dc27339af31bd0854bcbf47@relay/path",
			client: &ClientStub{
				resolveWithFiltersError: fmt.Errorf("failed to resolve with filters"),
			},
			expectedError: fmt.Errorf("failed to resolve with filters"),
			expectedCalledWith: struct {
				relays  []string
				filters nostr.Filters
			}{
				relays: []string{"ws://relay"},
				filters: nostr.Filters{{
					Kinds:   []int{30617},
					Authors: []string{"96beaaaeffcf4ca1b1e988bc5f112602b20945b67dc27339af31bd0854bcbf47"},
					Tags:    nostr.TagMap{"d": []string{"path"}},
				}},
			},
		},
		{
			name:      "fail to resolve with nip05",
			remoteRaw: "nostr://nip05@relay.str/path",
			client: &ClientStub{
				resolveWithNip05Error: fmt.Errorf("failed to resolve with nip05"),
			},
			expectedError:      fmt.Errorf("failed to resolve with nip05"),
			expectedCalledWith: "nostr://nip05@relay.str/path",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			s := NewService(tc.client)

			remote, err := s.ResolveRemote(tc.remoteRaw)
			if tc.expectedError != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tc.expectedError.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedRemote, remote.String())
			}
			assert.Equal(t, tc.expectedCalledWith, tc.client.resolveCalledWith)
		})
	}
}

type ClientStub struct {
	IClient
	resolveCalledWith any

	resolveWithNip05Response string
	resolveWithNip05Error    error

	resolveWithFiltersResponse string
	resolveWithFiltersError    error
}

func (c *ClientStub) ResolveWithNip05(remote *domain.Remote) (*domain.Remote, error) {
	c.resolveCalledWith = remote.String()
	resp, _ := domain.ParseRemote(c.resolveWithNip05Response)
	return resp, c.resolveWithNip05Error
}

func (c *ClientStub) ResolveWithFilters(relays []string, filters nostr.Filters) (*domain.Remote, error) {
	c.resolveCalledWith = struct {
		relays  []string
		filters nostr.Filters
	}{relays, filters}
	resp, _ := domain.ParseRemote(c.resolveWithFiltersResponse)
	return resp, c.resolveWithFiltersError
}
