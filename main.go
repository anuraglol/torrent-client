package main

import (
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jackpal/bencode-go"
)

var (
	workQueue          chan *pieceWork
	results            chan *pieceResult
	totalPieces        int
	maxConcurrentPeers = 5
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: <program> <inpath>")
		return
	}
	inPath := os.Args[1]

	tf, err := openTorrentFile(inPath)
	if err != nil {
		log.Fatalf("error opening torrent file: %v", err)
	}

	peers, err := tf.getPeers()
	if err != nil {
		log.Fatalf("error getting peers: %v", err)
	}

	totalPieces = len(tf.PieceHashes)
	workQueue = make(chan *pieceWork, totalPieces)
	results = make(chan *pieceResult, totalPieces)

	for i, hash := range tf.PieceHashes {
		length := tf.PieceLength
		if i == totalPieces-1 {
			length = tf.Length % tf.PieceLength
			if length == 0 {
				length = tf.PieceLength
			}
		}
		workQueue <- &pieceWork{index: i, hash: hash, length: length}
	}

	peerChan := make(chan *Peer, maxConcurrentPeers)
	for range maxConcurrentPeers {
		go func() {
			for peer := range peerChan {
				peer.startWorker(tf, workQueue, results)
			}
		}()
	}

	go func() {
		for i := range peers {
			peerChan <- &peers[i]
		}
	}()

	buf := make([][]byte, totalPieces)
	donePieces := 0
	for donePieces < totalPieces {
		res := <-results
		if !verifyPiece(tf, res.index, res.buf) {
			workQueue <- &pieceWork{index: res.index, hash: tf.PieceHashes[res.index], length: len(res.buf)}
			continue
		}
		buf[res.index] = res.buf
		donePieces++

		percent := float64(donePieces) / float64(totalPieces) * 100
		fmt.Printf("\r(%.2f%%) downloaded piece #%d\n", percent, res.index)
	}
	fmt.Println("\ndownload completed")
	close(workQueue)

	assembleFile(tf.Name, buf)
}

func openTorrentFile(path string) (*torrentFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bto := bencodeTorrent{}
	err = bencode.Unmarshal(file, &bto)
	if err != nil {
		return nil, err
	}
	return bto.toTorrentFile()
}

func (tf *torrentFile) getPeers() ([]Peer, error) {
	var peerID [20]byte
	_, err := rand.Read(peerID[:])
	if err != nil {
		return nil, err
	}
	resp, err := tf.requestTracker(peerID, 6881)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	interval, peers, err := parseTrackerResponse(resp.Body)
	if err != nil {
		return nil, err
	}
	log.Printf("tracker interval %d", interval)
	return peers, nil
}

func assembleFile(name string, buf [][]byte) {
	_ = os.MkdirAll("output", os.ModePerm)
	outPath := filepath.Join("output", name)
	outFile, err := os.Create(outPath)
	if err != nil {
		log.Fatalf("failed to create output file: %v", err)
	}
	defer outFile.Close()

	for _, b := range buf {
		_, err := outFile.Write(b)
		if err != nil {
			log.Fatalf("failed to write to output file: %v", err)
		}
	}
	fmt.Printf("file assembled successfully to %s\n", outPath)
}

func verifyPiece(tf *torrentFile, index int, data []byte) bool {
	hash := sha1.Sum(data)
	return string(hash[:]) == string(tf.PieceHashes[index][:])
}
