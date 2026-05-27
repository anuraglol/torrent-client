package main

import (
	"fmt"
	"net"
)

type bencodeInfo struct {
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
}

type bencodeTorrent struct {
	Announce string      `bencode:"announce"`
	Info     bencodeInfo `bencode:"info"`
}

type torrentFile struct {
	Announce    string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

type Peer struct {
	IP   net.IP
	Port uint16
}

func (p Peer) String() string {
	return fmt.Sprintf("%s:%d", p.IP, p.Port)
}

type pieceWork struct {
	index  int
	hash   [20]byte
	length int
}

type pieceResult struct {
	index int
	buf   []byte
}

type pieceProgress struct {
	work       *pieceWork
	index      int
	buf        []byte
	downloaded int
	requested  int
	backlog    int
}
