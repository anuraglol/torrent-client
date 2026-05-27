package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"
)

type Handshake struct {
	Pstr     string
	InfoHash [20]byte
	PeerID   [20]byte
}

func (h *Handshake) Serialize() []byte {
	buf := make([]byte, 1+len(h.Pstr)+8+20+20)
	buf[0] = byte(len(h.Pstr))
	curr := 1
	curr += copy(buf[curr:], h.Pstr)
	curr += copy(buf[curr:], make([]byte, 8))
	curr += copy(buf[curr:], h.InfoHash[:])
	curr += copy(buf[curr:], h.PeerID[:])
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

	copy(infoHash[:], handshakeBuf[pstrlen+8:pstrlen+8+20])
	copy(peerID[:], handshakeBuf[pstrlen+8+20:])

	return &Handshake{
		Pstr:     string(handshakeBuf[0:pstrlen]),
		InfoHash: infoHash,
		PeerID:   peerID,
	}, nil
}

func (p *Peer) ConnectAndHandshake(infoHash [20]byte, peerID [20]byte) (net.Conn, error) {
	address := net.JoinHostPort(p.IP.String(), strconv.Itoa(int(p.Port)))
	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		return nil, err
	}

	req := Handshake{
		Pstr:     "BitTorrent protocol",
		InfoHash: infoHash,
		PeerID:   peerID,
	}

	_, err = conn.Write(req.Serialize())
	if err != nil {
		conn.Close()
		return nil, err
	}

	res, err := ReadHandshake(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	if !bytes.Equal(res.InfoHash[:], infoHash[:]) {
		conn.Close()
		return nil, fmt.Errorf("info hash mismatch")
	}

	// fmt.Printf("successfully connected and handshaked with: %v\n", res.PeerID)

	return conn, nil
}
