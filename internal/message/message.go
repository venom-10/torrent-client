package message

import (
	"encoding/binary"
	"io"
)

type MessageID uint8

const (
	MsgChoked      MessageID = iota // peer will not send us piece data until unchoked
	MsgUnchoke                      // peer will now send us piece data - we may request
	MsgInterested                   // we want data from a peer who has pieces we lack
	MsgNotInterest                  // we no longer want anything from this peer
	MsgHave                         // INVENTORY: payload = 4-byte piece index - peer just gained this one piece
	MsgBitfield                     // INVENTORY: payload = bitfield - peer's full piece ownership, sent once after handshake
	MsgRequest                      // ASK: payload = index+begin+length (4+4+4) - please send me this one block
	MsgPiece                        // ANSWER: payload = index+begin+block - the actual file bytes, reply to a request (only message type carrying real data)
	MsgCancel                       // payload: index+begin+length - cancel a pending request (e.g. endgame mode)
)

type Message struct {
	ID      MessageID
	Payload []byte
}

func Read(r io.Reader) (*Message, error) {
	bufLen := make([]byte, 4)
	_, err := io.ReadFull(r, bufLen)
	if err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(bufLen)

	if length == 0 {
		return nil, nil
	}

	messageBuf := make([]byte, length)
	_, err = io.ReadFull(r, messageBuf)
	if err != nil {
		return nil, err
	}

	return &Message{
		ID:      MessageID(messageBuf[0]),
		Payload: messageBuf[1:],
	}, nil
}

func (m *Message) Serialize() []byte {
	if m == nil {
		return make([]byte, 4)
	}
	len := uint32(len(m.Payload)) + 1

	buf := make([]byte, len+4)
	binary.BigEndian.PutUint32(buf[0:4], len)
	buf[4] = byte(m.ID)
	copy(buf[5:], m.Payload)
	return buf
}

func FormatRequest(index, begin, length int) *Message {
	msg := make([]byte, 12)

	binary.BigEndian.PutUint32(msg[0:4], uint32(index))
	binary.BigEndian.PutUint32(msg[4:8], uint32(begin))
	binary.BigEndian.PutUint32(msg[8:12], uint32(length))

	return &Message{
		ID:      MsgRequest,
		Payload: msg,
	}

}
