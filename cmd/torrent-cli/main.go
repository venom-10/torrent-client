package main

import (
	"fmt"
	"log"
	"os"

	"github.com/venom-10/torrent-client/internal/metainfo"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run cmd/torrent-cli/main.go <path-to-torrent-file>")
	}

	filePath := os.Args[1]

	torrent, err := metainfo.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Announce URL: %s\n", torrent.Announce)
	fmt.Printf("File Name: %s\n", torrent.Info.Name)
	fmt.Printf("Length: %d bytes\n", torrent.Info.Length)
	fmt.Printf("Piece Length: %d bytes\n", torrent.Info.PieceLength)
	fmt.Printf("Total Pieces: %d\n", len(torrent.Info.Pieces))
}
