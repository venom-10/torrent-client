package main

import (
	"fmt"
	"log"
	"os"

	"github.com/venom-10/torrent-client/internal/download"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run cmd/torrent-cli/main.go <path-to-torrent-file>")
	}

	filePath := os.Args[1]

	if err := download.Run(filePath); err != nil {
		log.Fatal(err)
	}

	fmt.Println("download complete")
}
