package main

import (
	"fmt"
	"net"
	"time"
)

const (
	maxBacklog = 5
)

func (p *Peer) startWorker(tf *torrentFile, workQueue chan *pieceWork, results chan *pieceResult) {
	conn, err := net.DialTimeout("tcp", p.String(), 3*time.Second)
	if err != nil {
		fmt.Printf("could not connect to peer %s: %v\n", p.String(), err)
		return
	}
	defer conn.Close()

	_, err = completeHandshake(conn, tf.InfoHash, [20]byte{})
	if err != nil {
		fmt.Printf("handshake with peer %s failed: %v\n", p.String(), err)
		return
	}

	for pw := range workQueue {
		pp := newPieceProgress(pw)
		for pp.downloaded < pw.length {
			if !pp.isComplete() {
				// not complete, request a block
			}
		}
	}
}
