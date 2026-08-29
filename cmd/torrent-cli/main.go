package main

import (
	"github.com/venom-10/torrent-client/pkg/bencode"
	"os"
)

func main() {

	filePath := os.Args[1]
	bencode.OpenTorrent(filePath)
}
