package download

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/venom-10/torrent-client/internal/metainfo"
	"github.com/venom-10/torrent-client/internal/tracker"
)

type piece struct {
	index  int
	hash   [20]byte
	length int
}

type downloadedPiece struct {
	index int
	data  []byte
}

func getPieceLength(index int, tf *metainfo.TorrentFile) int {
	return min(tf.PieceLength, tf.Length-(index*tf.PieceLength))
}

func Run(filePath string) error {
	tf, err := metainfo.Open(filePath)
	if err != nil {
		return err
	}

	peerId, err := tracker.GeneratePeerId()
	if err != nil {
		return err
	}

	peerList, err := tracker.RequestPeers(tf, peerId)
	if err != nil {
		return err
	}
	if len(peerList) == 0 {
		return errors.New("tracker returned no peers")
	}

	numPieces := len(tf.PieceHashes)
	pending := make(chan piece, numPieces)
	finished := make(chan downloadedPiece, numPieces)

	for i := range numPieces {
		pending <- piece{index: i, hash: tf.PieceHashes[i], length: getPieceLength(i, tf)}
	}

	var wg sync.WaitGroup
	for _, pr := range peerList {
		wg.Add(1)
		go func() {
			defer wg.Done()
			downloadFromPeer(pr, peerId, tf.InfoHash, pending, finished)
		}()
	}

	// closes once every peer goroutine has returned
	allPeersGone := make(chan struct{})
	go func() {
		wg.Wait()
		close(allPeersGone)
	}()

	if err := saveToFile(tf, finished, allPeersGone, numPieces); err != nil {
		return err
	}

	close(pending)
	return nil
}

func saveToFile(tf *metainfo.TorrentFile, finished <-chan downloadedPiece, allPeersGone <-chan struct{}, numPieces int) error {
	file, err := os.Create(tf.Name)
	if err != nil {
		return err
	}
	defer file.Close()

	for done := range numPieces {
		var dp downloadedPiece
		select {
		case dp = <-finished:
		case <-allPeersGone:
			// a peer may have sent its last piece right before exiting, so check once more
			select {
			case dp = <-finished:
			default:
				return fmt.Errorf("all peers disconnected, %d/%d pieces done", done, numPieces)
			}
		}

		offset := int64(dp.index) * int64(tf.PieceLength)
		if _, err := file.WriteAt(dp.data, offset); err != nil {
			return fmt.Errorf("writing piece %d: %w", dp.index, err)
		}

		fmt.Printf("piece %d done (%d/%d)\n", dp.index, done+1, numPieces)
	}

	return nil
}
