package main

import (
	"io"

	"github.com/jackpal/bencode-go"
)

type bencodeTrackerResp struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

func parseTrackerResponse(r io.Reader) (int, []Peer, error) {
	var tr bencodeTrackerResp
	err := bencode.Unmarshal(r, &tr)
	if err != nil {
		return 0, nil, err
	}

	peers, err := toPeer(tr.Peers)
	if err != nil {
		return 0, nil, err
	}

	return tr.Interval, peers, nil
}
