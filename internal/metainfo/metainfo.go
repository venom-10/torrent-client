package metainfo

import (
	"os"

	"github.com/venom-10/torrent-client/internal/bencode"
)

func Open(filePath string) (*bencode.BencodeTorrent, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return bencode.Parse(data)
}
