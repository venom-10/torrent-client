package tracker

import (
	"crypto/rand"
	// "fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/venom-10/torrent-client/internal/bencode"
	"github.com/venom-10/torrent-client/internal/metainfo"
	"github.com/venom-10/torrent-client/internal/peers"
)

func generatePeerId() ([20]byte, error) {
	var id [20]byte
	copy(id[:8], []byte("venom-10"))
	_, err := rand.Read(id[8:])
	return id, err
}

func buildTrackerURL(tf *metainfo.TorrentFile, peerId [20]byte) (string, error) {
	base, err := url.Parse(tf.Announce)
	if err != nil {
		return "", err
	}

	params := url.Values{
		"info_hash":  []string{string(tf.InfoHash[:])},
		"peer_id":    []string{string(peerId[:])},
		"port":       []string{"6881"},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(tf.Length)},
	}

	base.RawQuery = params.Encode()
	return base.String(), err
}

func RequestPeers(tf *metainfo.TorrentFile) ([]peers.Peer, error) {
	peerId, err := generatePeerId()
	if err != nil {
		return nil, err
	}
	url, err := buildTrackerURL(tf, peerId)
	if err != nil {
		return nil, err
	}

	c := &http.Client{Timeout: 15 * time.Second}

	resp, err := c.Get(url)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	tr, err := bencode.ParseTrackerResponse(data)

	if err != nil {
		return nil, err
	}

	return peers.Unmarshal([]byte(tr.Peers))
}
