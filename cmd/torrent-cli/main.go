package main

import (
	// "fmt"
	"fmt"
	"log"
	"os"

	"github.com/venom-10/torrent-client/internal/metainfo"
	"github.com/venom-10/torrent-client/internal/tracker"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run cmd/torrent-cli/main.go <path-to-torrent-file>")
	}

	filePath := os.Args[1]

	torrentFile, err := metainfo.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}

	peers, _ := tracker.RequestPeers(torrentFile)
	fmt.Println(peers)

}
