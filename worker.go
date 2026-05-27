package main

import (
	"encoding/binary"
	"fmt"
	"time"
)

const (
	maxBacklog = 5
)

func (p *Peer) startWorker(tf *torrentFile, workQueue chan *pieceWork, results chan *pieceResult) {
	var peerID [20]byte
	c, err := New(*p, peerID, tf.InfoHash)
	if err != nil {
		return
	}
	defer c.Conn.Close()

	err = c.SendInterested()
	if err != nil {
		fmt.Printf("failed to send interested to %s: %v\n", p.String(), err)
		return
	}

	for pw := range workQueue {
		if !c.Bitfield.HasPiece(pw.index) {
			workQueue <- pw
			continue
		}
		buf, err := attemptDownloadPiece(c, pw)
		if err != nil {
			fmt.Printf("failed to download piece %d from %s: %v\n", pw.index, p.String(), err)
			workQueue <- pw
			return
		}

		results <- &pieceResult{index: pw.index, buf: buf}
	}
}

func attemptDownloadPiece(c *Client, pw *pieceWork) ([]byte, error) {
	state := pieceProgress{
		index: pw.index,
		buf:   make([]byte, pw.length),
	}

	c.Conn.SetDeadline(time.Now().Add(15 * time.Second))
	defer c.Conn.SetDeadline(time.Time{})

	for state.downloaded < pw.length {
		if c.Choked {

			msg, err := c.Read()
			if err != nil {
				return nil, err
			}
			if msg == nil {
				continue
			}
			if msg.ID == MsgUnchoke {
				c.Choked = false
			}
			continue
		}

		blockSize := 16384
		if pw.length-state.requested < blockSize {
			blockSize = pw.length - state.requested
		}

		if state.requested < pw.length {
			err := c.SendRequest(pw.index, state.requested, blockSize)
			if err != nil {
				return nil, err
			}
			state.requested += blockSize
		}

		msg, err := c.Read()
		if err != nil {
			return nil, err
		}
		if msg == nil {
			continue
		}

		switch msg.ID {
		case MsgChoke:
			c.Choked = true
		case MsgUnchoke:
			c.Choked = false
		case MsgPiece:
			if len(msg.Payload) < 8 {
				return nil, fmt.Errorf("piece message payload too short: %d bytes", len(msg.Payload))
			}

			begin := int(binary.BigEndian.Uint32(msg.Payload[4:8]))
			blockData := msg.Payload[8:]

			if begin+len(blockData) > len(state.buf) {
				return nil, fmt.Errorf("peer sent block out of bounds: begin %d, length %d, max %d", begin, len(blockData), len(state.buf))
			}

			copy(state.buf[begin:], blockData)

			state.downloaded += len(blockData)
		}
	}

	return state.buf, nil
}
