package nostr

import (
	"context"
	"fmt"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip05"
	"github.com/nbd-wtf/go-nostr/nip19"
	"log"
	"os"
	"strings"
)

// GetCredentials tries to return a locally stored private key from your git credentials
func GetCredentials(userName, userEmail string) (string, string) {
	// if we have a pubkey, we try to find a matching pvt key
	if nostr.IsValidPublicKey(userName) {
		return findKeys(userName)
	}

	// if we have npub, we try to find a pvtkey locally
	prefix, decoded, err := nip19.Decode(userName)
	if err == nil && prefix == "npub" {
		return findKeys(decoded.(string))
	}

	// if we have a nip05 identifier, we try to find a pvtkey locally
	if nip05.IsValidIdentifier(userEmail) {
		resp, err := nip05.QueryIdentifier(context.Background(), userEmail)
		if err == nil {
			return findKeys(resp.PublicKey)
		}
	}

	// try to use userName to read credentials from a file
	pvtKey, pubKey := readKeys(userName)
	if pvtKey != "" && pubKey != "" {
		return pvtKey, pubKey
	}

	// try to use userEmail to read credentials from a file
	pvtKey, pubKey = readKeys(userEmail)
	if pvtKey != "" && pubKey != "" {
		return pvtKey, pubKey
	}

	// finally, try to read from the default key file
	pvtKey, pubKey = readKeys("key")
	if pvtKey != "" && pubKey != "" {
		return pvtKey, pubKey
	}

	return "", ""
}

func readKeys(keyName string) (string, string) {
	home, _ := os.UserHomeDir()
	keyRaw, _ := readNostrFile(fmt.Sprintf("%s/.nostr/%s", home, keyName), "")
	if keyRaw != nil {
		pubKey, err := nostr.GetPublicKey(string(keyRaw))
		if err != nil {
			log.Fatal(err)
		}
		return string(keyRaw), pubKey
	}

	return "", ""
}

func findKeys(key string) (string, string) {
	// iterate over .pub files in ~/.nostr
	home, _ := os.UserHomeDir()
	files, _ := os.ReadDir(fmt.Sprintf("%s/.nostr", home))
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.HasSuffix(file.Name(), ".pub") {
			pvtKey, pubKey := readKeys(strings.TrimSuffix(file.Name(), ".pub"))
			if pubKey == key && pvtKey != "" {
				return pvtKey, pubKey
			}
		}
	}

	return "", ""
}
