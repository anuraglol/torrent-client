package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"
)

var (
	progressMu sync.Mutex
	Downloaded []bool
)

func InitProgress(totalPieces int) {
	Downloaded = make([]bool, totalPieces)
}

func (p *Peer) StartPeerSession(conn net.Conn) error {
	defer conn.Close()

	p.Choked = true
	p.Interested = false

	timeout := 15 * time.Second
	conn.SetReadDeadline(time.Now().Add(timeout))

	msg, err := ReadMessage(conn)
	if err != nil {
		return err
	}

	if msg != nil && msg.ID == MsgBitfield {
		peerBitfield := msg.Payload
		_ = peerBitfield
	}

	interestedMsg := Message{ID: MsgInterested}
	conn.SetWriteDeadline(time.Now().Add(timeout))
	_, err = conn.Write(interestedMsg.Serialize())
	if err != nil {
		return err
	}
	p.Interested = true

	for {
		conn.SetReadDeadline(time.Now().Add(timeout))
		msg, err := ReadMessage(conn)
		if err != nil {
			return err
		}

		if msg == nil {
			continue
		}

		fmt.Printf("message received: %v\n", msg.ID)

		switch msg.ID {
		case MsgChoke:
			p.Choked = true
			timeout = 15 * time.Second

		case MsgUnchoke:
			p.Choked = false
			timeout = 30 * time.Second

			err := p.DownloadNextBlock(conn)
			if err != nil {
				return err
			}

		case MsgPiece:
			timeout = 30 * time.Second
			fmt.Printf("download progress: %v\n", Downloaded)

			err := p.HandlePieceMessage(msg, conn)
			if err != nil {
				return err
			}
		}
	}
}

func (p *Peer) DownloadNextBlock(conn net.Conn) error {
	if p.Choked {
		return nil
	}

	var targetPiece uint32
	found := false

	progressMu.Lock()
	for i, done := range Downloaded {
		if !done {
			targetPiece = uint32(i)
			found = true
			break
		}
	}
	progressMu.Unlock()

	if !found {
		return nil
	}

	index := targetPiece
	begin := uint32(0)
	length := uint32(16384)

	reqMsg := Message{
		ID:      MsgRequest,
		Payload: FormatRequestPayload(index, begin, length),
	}

	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err := conn.Write(reqMsg.Serialize())
	return err
}

func (p *Peer) HandlePieceMessage(msg *Message, conn net.Conn) error {
	if len(msg.Payload) < 8 {
		return nil
	}

	index := binary.BigEndian.Uint32(msg.Payload[0:4])

	progressMu.Lock()
	if int(index) < len(Downloaded) && !Downloaded[index] {
		Downloaded[index] = true
		fmt.Printf("Piece %d saved in memory!\n", index)
	}
	progressMu.Unlock()

	return p.DownloadNextBlock(conn)
}

func FormatRequestPayload(index, begin, length uint32) []byte {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], index)
	binary.BigEndian.PutUint32(payload[4:8], begin)
	binary.BigEndian.PutUint32(payload[8:12], length)
	return payload
}
