package bencode

type BencodeInfo struct {
	Pieces      string
	PieceLength int
	Length      int
	Name        string
}

type BencodeTorrent struct {
	Announce string
	Info     BencodeInfo
}
