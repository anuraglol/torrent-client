package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"github.com/jackpal/bencode-go"
)

func (bto *bencodeTorrent) toTorrentFile() (*torrentFile, error) {
	infoHash, err := bto.Info.hash()
	if err != nil {
		return nil, err
	}
	pieceHashes, err := bto.Info.splitPieceHashes()
	if err != nil {
		return nil, err
	}
	tf := torrentFile{
		Announce:    bto.Announce,
		InfoHash:    infoHash,
		PieceHashes: pieceHashes,
		PieceLength: bto.Info.PieceLength,
		Length:      bto.Info.Length,
		Name:        bto.Info.Name,
	}
	return &tf, nil
}

func (i *bencodeInfo) hash() ([20]byte, error) {
	var buf bytes.Buffer
	err := bencode.Marshal(&buf, *i)
	if err != nil {
		return [20]byte{}, err
	}
	h := sha1.Sum(buf.Bytes())
	return h, nil
}

func (i *bencodeInfo) splitPieceHashes() ([][20]byte, error) {
	hashLen := 20
	buf := []byte(i.Pieces)
	if len(buf)%hashLen != 0 {
		err := fmt.Errorf("received malformed pieces of length %d", len(buf))
		return nil, err
	}
	numHashes := len(buf) / hashLen
	hashes := make([][20]byte, numHashes)
	for i := 0; i < numHashes; i++ {
		copy(hashes[i][:], buf[i*hashLen:(i+1)*hashLen])
	}
	return hashes, nil
}

func (tf *torrentFile) requestTracker(peerID [20]byte, port uint16) (*http.Response, error) {
	base, err := url.Parse(tf.Announce)
	if err != nil {
		return nil, err
	}
	params := url.Values{
		"info_hash":  []string{string(tf.InfoHash[:])},
		"peer_id":    []string{string(peerID[:])},
		"port":       []string{strconv.Itoa(int(port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(tf.Length)},
	}
	base.RawQuery = params.Encode()
	resp, err := http.Get(base.String())
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func toPeer(peers string) ([]Peer, error) {
	peerSize := 6
	buf := []byte(peers)
	if len(buf)%peerSize != 0 {
		return nil, fmt.Errorf("invalid peers field length")
	}
	numPeers := len(buf) / peerSize
	res := make([]Peer, numPeers)
	for i := 0; i < numPeers; i++ {
		offset := i * peerSize
		res[i].IP = net.IP(buf[offset : offset+4])
		res[i].Port = binary.BigEndian.Uint16(buf[offset+4 : offset+6])
	}
	return res, nil
}
