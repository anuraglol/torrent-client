package main

import (
	"encoding/binary"
	"io"
	"net"

	"github.com/jackpal/bencode-go"
)

type bencodeTrackerResponse struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

type Peer struct {
	IP         net.IP
	Port       uint16
	Choked     bool
	Interested bool
}

func parseTrackerResponse(r io.Reader) (int, []Peer, error) {
	var tr bencodeTrackerResponse
	err := bencode.Unmarshal(r, &tr)
	if err != nil {
		return 0, nil, err
	}

	peersBytes := []byte(tr.Peers)

	var peers []Peer
	for i := 0; i < len(peersBytes); i += 6 {
		if i+6 > len(peersBytes) {
			break
		}

		peer := Peer{
			IP:   net.IP(peersBytes[i : i+4]),
			Port: binary.BigEndian.Uint16(peersBytes[i+4 : i+6]),
		}
		peers = append(peers, peer)
	}

	return tr.Interval, peers, nil
}
