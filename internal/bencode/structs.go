package bencode

type BencodeInfo struct {
	Pieces      [][20]byte
	PieceLength int
	Length      int
	Name        string
}

type BencodeTorrent struct {
	Announce string
	Info     BencodeInfo
	InfoHash [20]byte
}

type TrackerResp struct {
	Interval int64
	Peers    string
}
