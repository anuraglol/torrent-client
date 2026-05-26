package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

const MaxBlockSize = 16384

func DownloadPiece(conn net.Conn, index int, pieceLength int) ([]byte, error) {
	conn.SetDeadline(time.Now().Add(30 * time.Second))
	defer conn.SetDeadline(time.Time{})

	pieceBuf := make([]byte, pieceLength)
	requested := 0
	downloaded := 0

	for downloaded < pieceLength {
		for requested < pieceLength {
			blockSize := MaxBlockSize
			if pieceLength-requested < blockSize {
				blockSize = pieceLength - requested
			}

			payload := make([]byte, 12)
			binary.BigEndian.PutUint32(payload[0:4], uint32(index))
			binary.BigEndian.PutUint32(payload[4:8], uint32(requested))
			binary.BigEndian.PutUint32(payload[8:12], uint32(blockSize))

			reqMsg := Message{
				ID:      MsgRequest,
				Payload: payload,
			}

			_, err := conn.Write(reqMsg.Serialize())
			if err != nil {
				return nil, err
			}
			requested += blockSize
		}

		msg, err := ReadMessage(conn)
		if err != nil {
			return nil, err
		}

		if msg == nil {
			continue
		}

		if msg.ID == MsgPiece {
			if len(msg.Payload) < 8 {
				return nil, fmt.Errorf("piece payload too short")
			}

			resIndex := binary.BigEndian.Uint32(msg.Payload[0:4])
			resBegin := binary.BigEndian.Uint32(msg.Payload[4:8])
			blockData := msg.Payload[8:]

			if int(resIndex) != index {
				return nil, fmt.Errorf("expected piece %d, got %d", index, resIndex)
			}

			if int(resBegin)+len(blockData) > pieceLength {
				return nil, fmt.Errorf("block offset out of bounds")
			}

			copy(pieceBuf[resBegin:], blockData)
			downloaded += len(blockData)
		}
	}

	return pieceBuf, nil
}
