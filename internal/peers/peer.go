package peers

import (
	"encoding/binary"
	"errors"
	"net"
)

type Peer struct {
	IP   net.IP
	Port uint16
}

func Unmarshal(peerBytes []byte) ([]Peer, error) {
	peerSize := 6
	totPeers := len(peerBytes) / peerSize

	if len(peerBytes)%peerSize != 0 {
		return nil, errors.New("received malformed peers")
	}

	peers := make([]Peer, totPeers)
	for i := range peers {
		peerStart := i * peerSize
		peers[i].IP = net.IP(peerBytes[peerStart : peerStart+4])
		peers[i].Port = binary.BigEndian.Uint16(peerBytes[peerStart+4 : peerStart+6])
	}

	return peers, nil
}
