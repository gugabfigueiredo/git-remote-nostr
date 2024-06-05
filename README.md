# git-remote-nostr - WIP

A git remote-helper to add support for nostr remote urls.

## Roadmap

- [x] create git-remote-helper
- [x] clone from remote url
- [x] fetch from remote url
- [x] push to remote url
- [x] parse remote url
- [ ] fetch remote url from nostr as necessary
- [ ] add support for single use ssh key pair as available/required by host

## Installation

```sh
$ go install github.com/gugabfigueiredo/git-remote-nostr
```

## Usage

```sh
$ git clone nostr::<remoteurl>
```