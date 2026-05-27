package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"time"
)

type Handshake struct {
	Pstr     string
	InfoHash [20]byte
	PeerID   [20]byte
}

func NewHandshake(infoHash, peerID [20]byte) []byte {
	pstr := "BitTorrent protocol"
	buf := make([]byte, len(pstr)+49)
	buf[0] = byte(len(pstr))
	curr := 1
	curr += copy(buf[curr:], pstr)
	curr += copy(buf[curr:], make([]byte, 8))
	curr += copy(buf[curr:], infoHash[:])
	curr += copy(buf[curr:], peerID[:])
	return buf
}

func ReadHandshake(r io.Reader) (*Handshake, error) {
	lengthBuf := make([]byte, 1)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, err
	}
	pstrlen := int(lengthBuf[0])

	if pstrlen == 0 {
		return nil, fmt.Errorf("pstrlen cannot be 0")
	}

	handshakeBuf := make([]byte, pstrlen+8+20+20)
	_, err = io.ReadFull(r, handshakeBuf)
	if err != nil {
		return nil, err
	}

	var infoHash [20]byte
	var peerID [20]byte

	pstr := string(handshakeBuf[0:pstrlen])
	copy(infoHash[:], handshakeBuf[pstrlen+8:pstrlen+8+20])
	copy(peerID[:], handshakeBuf[pstrlen+8+20:])

	return &Handshake{
		Pstr:     pstr,
		InfoHash: infoHash,
		PeerID:   peerID,
	}, nil
}

func (h *Handshake) validate(infoHash [20]byte) bool {
	return bytes.Equal(h.InfoHash[:], infoHash[:])
}

func completeHandshake(conn net.Conn, infohash, peerID [20]byte) (*Handshake, error) {
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{})

	_, err := conn.Write(NewHandshake(infohash, peerID))
	if err != nil {
		return nil, err
	}

	res, err := ReadHandshake(conn)
	if err != nil {
		return nil, err
	}
	if !res.validate(infohash) {
		return nil, fmt.Errorf("expected infohash %x but got %x", infohash, res.InfoHash)
	}
	return res, nil
}
