package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/jackpal/bencode-go"
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

type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

func calculateSHA1(data interface{}) ([]byte, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	h := sha1.New()
	h.Write(bytes)
	return h.Sum(nil), nil
}

func (bto bencodeTorrent) toTorrentFile() (TorrentFile, error) {
	tf := TorrentFile{}
	tf.Announce, tf.Name, tf.Length, tf.PieceLength = bto.Announce, bto.Info.Name, bto.Info.Length, bto.Info.PieceLength

	infoHash, err := calculateSHA1(bto.Info)
	if err == nil {
		tf.InfoHash = [20]byte(infoHash)
	}

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

	return tf, nil
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: <program> <inPath> <outPath>")
		return
	}

	inPath := os.Args[1]

	file, err := os.Open(inPath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	bto := bencodeTorrent{}
	er := bencode.Unmarshal(file, &bto)
	if er != nil {
		log.Fatalf("oops, error captured: %v", er.Error())
		return
	} else {
		fmt.Printf("%v", &bto.Announce)
	}

	tf, err := bto.toTorrentFile()
	if err != nil {
		log.Fatalf("oops, error captured: %v", er.Error())
		return
	} else {
		fmt.Printf("file name: %v", tf.Name)
	}
}
