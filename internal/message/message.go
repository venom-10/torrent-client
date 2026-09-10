package message

import (
	"encoding/binary"
	"io"
)

type MessageID uint8

const (
	MsgChoked MessageID = iota
	MsgUnchoke
	MsgInterested
	MsgNotInterest
	MsgHave
	MsgBitfield
	MsgRequest
	MsgPiece
	MsgCancel
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
