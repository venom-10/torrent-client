# torrent-client

A BitTorrent client written from scratch in Go — no external dependencies, no
third-party torrent/bencode libraries. Built as a learning project to understand
the BitTorrent protocol end to end: bencode encoding, tracker communication, the
peer wire protocol, and concurrent piece downloading.

## What it does

Given a `.torrent` file, it:
1. Parses the file's bencode-encoded metadata and computes the info hash.
2. Announces to the torrent's tracker over HTTP(S) and gets back a list of peers.
3. Connects to multiple peers concurrently, completes the BitTorrent handshake
   with each, and reads their piece bitfields.
4. Downloads pieces in parallel across all connected peers — each piece is split
   into 16KB blocks and requested with a pipelined backlog for throughput.
5. Verifies every downloaded piece against its SHA1 hash from the torrent
   metadata before writing it to disk at the correct offset.

## Usage

```sh
go run ./cmd/torrent-cli <path-to-torrent-file>
```

The downloaded file is written to the current directory, named after the file
name specified in the torrent's metadata.

## Project layout

```
cmd/torrent-cli/       CLI entry point
internal/
  bencode/              bencode decoder + .torrent/tracker-response parsing
  metainfo/              .torrent file -> TorrentFile struct
  tracker/               HTTP(S) tracker announce, peer discovery
  peers/                 compact peer-list parsing
  handshake/              BitTorrent handshake (protocol string, info hash, peer ID)
  message/               peer wire protocol message framing (choke/unchoke/have/
                          bitfield/request/piece/cancel)
  bitfield/               per-peer piece-ownership bitfield
  p2p/                    per-connection client: handshake, message handling,
                          choke/interest state, pipelined block downloading
  download/               concurrent multi-peer orchestration: work queue, worker
                          goroutines, single-writer disk output
```

## Known limitations

- **Single-file torrents only** — no support for multi-file torrents yet.
- **HTTP(S) trackers only** — no UDP tracker support.
- **Compact peer format only** — trackers that return the non-compact
  (dictionary-list) peer format, common for IPv6-heavy swarms, aren't handled yet.
- **No peer-pool replenishment** — if every currently-connected peer disconnects
  mid-download (e.g. a brief network interruption), the download fails rather than
  re-announcing to the tracker for a fresh peer list.
- **No resume support** — a download can't be resumed across separate runs; a
  partially-written output file from an interrupted run is not reused.
- **Sequential piece assignment** — pieces are handed out in order, not
  rarest-first.
- **No automated tests yet.**

## Requirements

Go 1.23.4+. No external dependencies — everything (bencode, HTTP, SHA1, TCP) is
standard library.
