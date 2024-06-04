# git-remote-nostr - WIP

A git remote-helper to host a remote anywhere and publish in nostr.

## Roadmap

- [x] create git-remote-helper
- [x] clone from remote url
- [x] push to remote url
- [x] fetch from remote url
- [ ] create nostr client
- [ ] parse remote url
- [ ] fetch remote url from nostr if necessary

## Installation

```sh
$ go install github.com/gugabfigueiredo/git-remote-nostr
```

## Usage

```sh
$ git clone nostr::<remoteurl>
```