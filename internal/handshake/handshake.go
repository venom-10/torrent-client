package handshake

import (
	"errors"
	"fmt"
)

type Handshake struct {
	Pstr     string
	InfoHash [20]byte
	PeerId   [20]byte
}

func NewHandshake(infoHash [20]byte, peerId [20]byte) *Handshake {
	return &Handshake{
		Pstr:     "BitTorrent protocol",
		InfoHash: infoHash,
		PeerId:   peerId,
	}
}

func (h *Handshake) Serialize() []byte {
	buf := make([]byte, 68)

	buf[0] = byte(len(h.Pstr))

	nxt := 1
	nxt += copy(buf[1:], []byte(h.Pstr))
	nxt += 8
	nxt += copy(buf[nxt:], h.InfoHash[:])
	copy(buf[nxt:], h.PeerId[:])

	return buf
}

func ParseHandshake(b []byte) (*Handshake, error) {
	if len(b) < 68 {
		return nil, errors.New("handshake too short")
	}

	protoLen := int(b[0])
	if protoLen != 19 {
		return nil, fmt.Errorf("invalid protocol length: %d", protoLen)
	}

	pstr := string(b[1:20])
	if pstr != "BitTorrent protocol" {
		return nil, fmt.Errorf("unrecognized protocol: %s", pstr)
	}

	var infoHash [20]byte
	copy(infoHash[:], b[28:48])

	var peerId [20]byte
	copy(infoHash[:], b[48:68])

	return &Handshake{
		Pstr:     pstr,
		InfoHash: infoHash,
		PeerId:   peerId,
	}, nil

}
