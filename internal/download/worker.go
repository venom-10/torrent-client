package download

import (
	"crypto/sha1"
	"log"
	"time"

	"github.com/venom-10/torrent-client/internal/p2p"
	"github.com/venom-10/torrent-client/internal/peers"
)

func downloadFromPeer(p peers.Peer, peerId [20]byte, infoHash [20]byte, pending chan piece, finished chan<- downloadedPiece) {
	client, err := p2p.NewClient(p, peerId, infoHash)

	if err != nil {
		log.Printf("peer %s: %v", p.FormatIp(), err)
		return
	}
	defer client.Conn.Close()

	if err := client.SendInterested(); err != nil {
		log.Printf("peer %s: %v", p.FormatIp(), err)
		return
	}

	if err := client.WaitForUnchoke(10 * time.Second); err != nil {
		log.Printf("peer %s: %v", p.FormatIp(), err)
		return
	}

	for pc := range pending {
		if !client.Bitfield.HasPiece(pc.index) {
			pending <- pc
			continue
		}

		buf, err := client.DownloadPiece(pc.index, pc.length)

		if err != nil {
			log.Printf("Downloading fail from peer %s piece %d: %v", p.FormatIp(), pc.index, err)
			pending <- pc
			return
		}

		dpc := downloadedPiece{
			index: pc.index,
			data:  buf,
		}

		if pc.hash != sha1.Sum(dpc.data) {
			log.Printf("Downloaded piece %d from peer %s failed hash check", pc.index, p.FormatIp())
			pending <- pc
			continue
		}
		finished <- dpc
	}

}
