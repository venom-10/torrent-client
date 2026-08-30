package metainfo

import (
	"os"

	"github.com/venom-10/torrent-client/internal/bencode"
)

type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

func Open(filePath string) (*TorrentFile, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	bt, err := bencode.Parse(data)

	tf := &TorrentFile{
		Announce:    bt.Announce,
		InfoHash:    bt.InfoHash,
		PieceHashes: bt.Info.Pieces,
		PieceLength: bt.Info.PieceLength,
		Length:      bt.Info.Length,
		Name:        bt.Info.Name,
	}

	return tf, err
}
