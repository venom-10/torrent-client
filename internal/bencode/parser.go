package bencode

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"strconv"
)

type parser struct {
	data    []byte
	pointer int
}

func (p *parser) parseInt() (int64, error) {
	start := p.pointer + 1
	idx := bytes.IndexByte(p.data[start:], 'e')
	if idx == -1 {
		return 0, errors.New("missing ending 'e' marker for integer")
	}
	end := start + idx
	intStr := string(p.data[start:end])
	val, err := strconv.ParseInt(intStr, 10, 64)
	if err != nil {
		return 0, err
	}

	p.pointer = end + 1
	return val, nil
}

func (p *parser) parseString() (string, error) {
	start := p.pointer

	idx := bytes.IndexByte(p.data[start:], ':')
	if idx == -1 {
		return "", errors.New("missing ':' marker for string")
	}
	colonIndex := start + idx

	stringLength, err := strconv.ParseInt(string(p.data[start:colonIndex]), 10, 64)
	if err != nil {
		return "", errors.New("can't convert string length to int")
	}

	strStart := colonIndex + 1
	strEnd := strStart + int(stringLength)

	if strEnd > len(p.data) {
		return "", errors.New("string length exceeds data bounds")
	}

	p.pointer = strEnd
	return string(p.data[strStart:strEnd]), nil
}

func (p *parser) parseList() ([]any, error) {

	p.pointer++ // Skip 'l' marker
	var list []any

	for p.pointer < len(p.data) && p.data[p.pointer] != 'e' {
		val, err := p.parseNext()
		if err != nil {
			return nil, err
		}
		list = append(list, val)
	}

	if p.pointer >= len(p.data) {
		return nil, errors.New("missing ending 'e' marker for list")
	}

	p.pointer++ // Skip 'e' marker
	return list, nil
}

func (p *parser) parseDict() (map[string]any, error) {
	p.pointer++ // Skip 'd' marker
	dict := make(map[string]any)

	for p.pointer < len(p.data) && p.data[p.pointer] != 'e' {
		// it can't start without number
		if p.data[p.pointer] < '0' || p.data[p.pointer] > '9' {
			return nil, errors.New("dictionary key must be a string")
		}

		key, err := p.parseString()
		if err != nil {
			return nil, err
		}

		valStart := p.pointer
		val, err := p.parseNext()
		if err != nil {
			return nil, err
		}
		valEnd := p.pointer

		dict[key] = val

		if key == "info" {
			dict["raw_info_bytes"] = p.data[valStart:valEnd]
		}
	}

	if p.pointer >= len(p.data) {
		return nil, errors.New("missing ending 'e' marker for dictionary")
	}

	p.pointer++ // Skip 'e' marker
	return dict, nil
}

func (p *parser) parseNext() (any, error) {
	if p.pointer >= len(p.data) {
		return nil, errors.New("unexpected end of data")
	}
	switch p.data[p.pointer] {
	case 'i':
		return p.parseInt()
	case 'l':
		return p.parseList()
	case 'd':
		return p.parseDict()
	default:
		if p.data[p.pointer] >= '0' && p.data[p.pointer] <= '9' {
			return p.parseString()
		}
		return nil, fmt.Errorf("unknown type at position %d: %c", p.pointer, p.data[p.pointer])
	}
}

func splitPieceHashes(pieces string) ([][20]byte, error) {
	hashLen := 20
	if len(pieces)%hashLen != 0 {
		return nil, fmt.Errorf("malformed pieces string: length is not a multiple of 20")
	}

	numHashes := len(pieces) / hashLen
	hashes := make([][20]byte, numHashes)

	for i := 0; i < numHashes; i++ {
		start := i * hashLen
		end := start + hashLen
		copy(hashes[i][:], pieces[start:end])
	}

	return hashes, nil
}

func unmarshalTorrent(parsedData any) (*BencodeTorrent, error) {
	root, ok := parsedData.(map[string]any)
	if !ok {
		return nil, errors.New("invalid torrent file: root is not a dictionary")
	}

	torrent := &BencodeTorrent{}

	if announce, ok := root["announce"].(string); ok {
		torrent.Announce = announce
	}

	if infoRaw, ok := root["info"].(map[string]any); ok {
		if name, ok := infoRaw["name"].(string); ok {
			torrent.Info.Name = name
		}
		if pieces, ok := infoRaw["pieces"].(string); ok {
			pieceHash, err := splitPieceHashes(pieces)
			if err != nil {
				return nil, errors.New("invalid pieceHashes")
			}
			torrent.Info.Pieces = pieceHash
		}
		if pieceLength, ok := infoRaw["piece length"].(int64); ok {
			torrent.Info.PieceLength = int(pieceLength)
		}
		if length, ok := infoRaw["length"].(int64); ok {
			torrent.Info.Length = int(length)
		}
	} else {
		return nil, errors.New("invalid torrent file: missing info dictionary")
	}

	if rawInfo, ok := root["raw_info_bytes"].([]byte); ok {
		torrent.InfoHash = sha1.Sum(rawInfo)
	} else {
		return nil, errors.New("invalid torrent file")
	}

	return torrent, nil
}

func unmarshalTrackerResponse(parsedData map[string]any) (*TrackerResp, error) {
	tr := &TrackerResp{}
	if interval, ok := parsedData["interval"].(int64); ok {
		tr.Interval = interval
	} else {
		return nil, errors.New("invalid tracker response: missing interval")
	}

	if peers, ok := parsedData["peers"].(string); ok {
		tr.Peers = peers
	} else {
		return nil, errors.New("try again: missing peers")
	}

	return tr, nil
}

func Parse(data []byte) (*BencodeTorrent, error) {
	p := &parser{
		data:    data,
		pointer: 0,
	}

	parsedData, err := p.parseNext()
	if err != nil {
		return nil, err
	}

	return unmarshalTorrent(parsedData)
}
func ParseTrackerResponse(data []byte) (*TrackerResp, error) {
	p := &parser{
		data:    data,
		pointer: 0,
	}

	parsedData, err := p.parseDict()
	if err != nil {
		return nil, err
	}

	return unmarshalTrackerResponse(parsedData)
}
