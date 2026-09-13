package p2p

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/venom-10/torrent-client/internal/handshake"
	"github.com/venom-10/torrent-client/internal/peers"
)

func ConnectToPeer(p peers.Peer, peerId [20]byte, infoHash [20]byte) (net.Conn, error) {
	addr := p.FormatIp()
	conn, err := net.DialTimeout("tcp", addr, 3*time.Second)

	if err != nil {
		return nil, err
	}

	if err := conn.SetDeadline(time.Now().Add(7 * time.Second)); err != nil {
		conn.Close()
		return nil, err
	}

	req := handshake.NewHandshake(infoHash, peerId)
	_, err = conn.Write(req.Serialize())
	if err != nil {
		conn.Close()
		return nil, err
	}

	resBuf := make([]byte, 68)
	_, err = io.ReadFull(conn, resBuf)
	if err != nil {
		conn.Close()
		return nil, err
	}

	resHandshake, err := handshake.ParseHandshake(resBuf)

	if err != nil {
		conn.Close()
		return nil, err
	}

	if resHandshake.InfoHash != infoHash {
		conn.Close()
		return nil, fmt.Errorf("expected infohash %x but got %x", infoHash, resHandshake.InfoHash)
	}

	if err := conn.SetDeadline(time.Time{}); err != nil {
		conn.Close()
		return nil, err
	}

	return conn, nil
}
