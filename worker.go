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
	// 1. Initialize the client (Handshake + receive Bitfield)
	var peerID [20]byte // You can pass your generated peerID here if needed
	c, err := New(*p, peerID, tf.InfoHash)
	if err != nil {
		// fmt.Printf("could not connect to peer %s: %v\n", p.String(), err)
		return
	}
	defer c.Conn.Close()

	// 2. Tell the peer we want data
	err = c.SendInterested()
	if err != nil {
		fmt.Printf("failed to send interested to %s: %v\n", p.String(), err)
		return
	}

	for pw := range workQueue {
		// If the peer doesn't have the piece we need, put it back and skip
		if !c.Bitfield.HasPiece(pw.index) {
			workQueue <- pw
			continue
		}
		buf, err := attemptDownloadPiece(c, pw)
		if err != nil {
			fmt.Printf("failed to download piece %d from %s: %v\n", pw.index, p.String(), err)
			workQueue <- pw // Put work back on queue to retry
			return          // Drop connection on network failure
		}

		// Send results back to main loop
		results <- &pieceResult{index: pw.index, buf: buf}
	}
}

func attemptDownloadPiece(c *Client, pw *pieceWork) ([]byte, error) {
	state := pieceProgress{
		index: pw.index,
		buf:   make([]byte, pw.length),
	}

	// Set a baseline deadline for network activity
	c.Conn.SetDeadline(time.Now().Add(15 * time.Second))
	defer c.Conn.SetDeadline(time.Time{})

	// Keep loop running until we fill our buffer
	for state.downloaded < pw.length {
		if c.Choked {
			// If choked, wait for messages until we get unchoked
			msg, err := c.Read()
			if err != nil {
				return nil, err
			}
			if msg == nil {
				continue // Keep-alive message
			}
			if msg.ID == MsgUnchoke {
				c.Choked = false
			}
			continue
		}

		// Cap block size to file size limit so we don't request out of bounds
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

		// Read incoming responses
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

			// Extract the 'begin' byte offset from bytes [4:8]
			begin := int(binary.BigEndian.Uint32(msg.Payload[4:8]))
			blockData := msg.Payload[8:]

			// Safety boundary checks to prevent out-of-bounds panics
			if begin+len(blockData) > len(state.buf) {
				return nil, fmt.Errorf("peer sent block out of bounds: begin %d, length %d, max %d", begin, len(blockData), len(state.buf))
			}

			// Copy the raw block data into our piece buffer at the correct offset
			copy(state.buf[begin:], blockData)

			// Advance progress counters
			state.downloaded += len(blockData)
		}
	}

	return state.buf, nil
}
