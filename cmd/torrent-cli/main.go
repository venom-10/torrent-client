package main

import (
	"crypto/sha1"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/venom-10/torrent-client/internal/metainfo"
	"github.com/venom-10/torrent-client/internal/p2p"
	"github.com/venom-10/torrent-client/internal/tracker"
)

const maxPeerAttempts = 5

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run cmd/torrent-cli/main.go <path-to-torrent-file>")
	}

	filePath := os.Args[1]

	torrentFile, err := metainfo.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}

	peerID, err := tracker.GeneratePeerId()
	if err != nil {
		log.Fatal(err)
	}

	peerList, err := tracker.RequestPeers(torrentFile, peerID)
	if err != nil {
		log.Fatal(err)
	}
	if len(peerList) == 0 {
		log.Fatal("tracker returned no peers")
	}

	attempts := min(len(peerList), maxPeerAttempts)

	var client *p2p.Client
	var connectedPeer string
	for i := 0; i < attempts; i++ {
		c, err := p2p.NewClient(peerList[i], peerID, torrentFile.InfoHash)
		if err != nil {
			log.Printf("peer %s: %v", peerList[i].FormatIp(), err)
			continue
		}
		client = c
		connectedPeer = peerList[i].FormatIp()
		break
	}
	if client == nil {
		log.Fatalf("could not establish a usable connection to any of %d peers tried", attempts)
	}
	defer client.Conn.Close()

	fmt.Printf("connected to %s\n", connectedPeer)

	if err := client.SendInterested(); err != nil {
		log.Fatal(err)
	}

	if err := client.WaitForUnchoke(10 * time.Second); err != nil {
		log.Fatal(err)
	}

	buf, err := client.DownloadPiece(0, torrentFile.PieceLength)

	if err != nil {
		log.Fatal("error downloading piece 0")
	}

	hash := sha1.Sum(buf)

	if hash != torrentFile.PieceHashes[0] {
		log.Fatal("wrong hashFile received")
	}

	file, err := os.Create("./deb_ubuntu")
	if err != nil {
		log.Fatal("error creating file")
	}
	defer file.Close()

	file.WriteAt(buf, 0)


	fmt.Println("peer unchoked us - ready to request blocks")
}
