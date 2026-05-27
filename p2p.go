package main

import (
	"encoding/binary"
	"fmt"
	"net"
)

type peerConn struct {
	conn net.Conn
}

func (pc *peerConn) requestBlock(index, begin, length int) error {
	req := Message{ID: MsgRequest}
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	binary.BigEndian.PutUint32(payload[8:12], uint32(length))
	req.Payload = payload
	_, err := pc.conn.Write(req.Serialize())
	return err
}

func (pc *peerConn) handlePiece(msg *Message, pp *pieceProgress) (int, error) {
	if msg.ID != MsgPiece {
		return 0, fmt.Errorf("expected piece message, got %d", msg.ID)
	}
	if len(msg.Payload) < 8 {
		return 0, fmt.Errorf("payload too short")
	}
	index := binary.BigEndian.Uint32(msg.Payload[0:4])
	begin := binary.BigEndian.Uint32(msg.Payload[4:8])
	data := msg.Payload[8:]
	if int(index) != pp.index {
		return 0, fmt.Errorf("expected piece index %d, got %d", pp.index, index)
	}
	if int(begin)+len(data) > len(pp.buf) {
		return 0, fmt.Errorf("data out of bounds")
	}
	copy(pp.buf[begin:], data)
	pp.downloaded += len(data)
	return len(data), nil
}

func newPieceProgress(work *pieceWork) *pieceProgress {
	return &pieceProgress{
		work:  work,
		index: work.index,
		buf:   make([]byte, work.length),
	}
}

func (pp *pieceProgress) isComplete() bool {
	return pp.downloaded == len(pp.buf)
}
