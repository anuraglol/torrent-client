package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/jackpal/bencode-go"
)

func calculateSHA1(btoInfo interface{}) ([20]byte, error) {
	var buf bytes.Buffer
	err := bencode.Marshal(&buf, btoInfo)
	if err != nil {
		return [20]byte{}, err
	}

	h := sha1.New()
	h.Write(buf.Bytes())

	var infoHash [20]byte
	copy(infoHash[:], h.Sum(nil))
	return infoHash, nil
}

func (bto bencodeTorrent) toTorrentFile() (*TorrentFile, error) {
	tf := TorrentFile{}
	tf.Announce, tf.Name, tf.Length, tf.PieceLength = bto.Announce, bto.Info.Name, bto.Info.Length, bto.Info.PieceLength

	infoHash, err := calculateSHA1(bto.Info)
	if err != nil {
		return nil, err
	}
	tf.InfoHash = (infoHash)

	data := []byte(bto.Info.Pieces)
	var pieceHases [][20]byte

	for i := 0; i < len(data); i += 20 {
		end := i + 20
		if end > len(data) {
			end = len(data)
		}

		var chunk [20]byte
		copy(chunk[:], data[i:end])
		pieceHases = append(pieceHases, chunk)
	}

	tf.PieceHashes = pieceHases

	return &tf, nil
}

func (t *TorrentFile) requestTracker(peerID [20]byte, port uint16) (*http.Response, error) {
	base, err := url.Parse(t.Announce)
	if err != nil {
		return nil, err
	}

	params := url.Values{
		"port":       []string{strconv.Itoa(int(port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(t.Length)},
	}

	infoHashEncoded := renderPercentQuery(t.InfoHash[:])
	peerIDEncoded := renderPercentQuery(peerID[:])

	rawQuery := fmt.Sprintf("info_hash=%s&peer_id=%s&%s", infoHashEncoded, peerIDEncoded, params.Encode())
	base.RawQuery = rawQuery

	fmt.Printf("url: %v\n", base.String())

	resp, err := http.Get(base.String())
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func renderPercentQuery(b []byte) string {
	var buf bytes.Buffer
	for _, val := range b {
		buf.WriteString(fmt.Sprintf("%%%02x", val))
	}
	return buf.String()
}
