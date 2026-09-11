package p2p

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"github.com/venom-10/torrent-client/internal/bitfield"
	"github.com/venom-10/torrent-client/internal/message"
	"github.com/venom-10/torrent-client/internal/peers"
)

type Client struct {
	Conn     net.Conn
	Bitfield bitfield.Bitfield

	AmChoking      bool
	AmInterested   bool
	PeerChoking    bool
	PeerInterested bool
}

func NewClient(peer peers.Peer, peerID, infoHash [20]byte) (*Client, error) {
	conn, err := ConnectToPeer(peer, peerID, infoHash)
	if err != nil {
		return nil, err
	}

	if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		conn.Close()
		return nil, err
	}

	msg, err := message.Read(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}
	if msg == nil {
		conn.Close()
		return nil, fmt.Errorf("expected bitfield, got keep-alive")
	}
	if msg.ID != message.MsgBitfield {
		conn.Close()
		return nil, fmt.Errorf("expected bitfield, got message ID %d", msg.ID)
	}

	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		conn.Close()
		return nil, err
	}

	return &Client{
		Conn:           conn,
		Bitfield:       bitfield.Bitfield(msg.Payload),
		AmChoking:      true,
		AmInterested:   false,
		PeerChoking:    true,
		PeerInterested: false,
	}, nil
}

func (c *Client) HandleMessage() (*message.Message, error) {
	msg, err := message.Read(c.Conn)
	if err != nil {
		return nil, err
	}
	if msg == nil {
		return nil, nil
	}

	switch msg.ID {
	case message.MsgChoked:
		c.PeerChoking = true
	case message.MsgUnchoke:
		c.PeerChoking = false
	case message.MsgInterested:
		c.PeerInterested = true
	case message.MsgNotInterest:
		c.PeerInterested = false
	case message.MsgHave:
		if len(msg.Payload) != 4 {
			return nil, fmt.Errorf("invalid have payload length %d", len(msg.Payload))
		}
		index := int(binary.BigEndian.Uint32(msg.Payload))
		c.Bitfield.SetPiece(index)
	case message.MsgBitfield:
		c.Bitfield = bitfield.Bitfield(msg.Payload)
	}

	return msg, nil
}


func (c *Client) SendInterested() error {
	msg := &message.Message{ID: message.MsgInterested}
	if _, err := c.Conn.Write(msg.Serialize()); err != nil {
		return err
	}
	c.AmInterested = true
	return nil
}

func (c *Client) WaitForUnchoke(timeout time.Duration) error {
	if err := c.Conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return err
	}
	defer c.Conn.SetReadDeadline(time.Time{})

	for c.PeerChoking {
		if _, err := c.HandleMessage(); err != nil {
			return fmt.Errorf("waiting for unchoke: %w", err)
		}
	}
	return nil
}
