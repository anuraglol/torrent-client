package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/jackpal/bencode-go"
)

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
		fmt.Printf("announce: %v\n", &bto.Announce)
	}

	tf, err := bto.toTorrentFile()
	if err != nil {
		log.Fatalf("oops, error captured: %v", er.Error())
		return
	} else {
		fmt.Printf("file name: %v\n", tf.Name)
	}

	token := make([]byte, 20)
	rand.Read(token)
	resp, e := tf.requestTracker([20]byte(token), 6881)
	if e != nil {
		log.Fatalf("oops, error captured: %v", e.Error())
		return
	}
	defer resp.Body.Close()

	interval, peers, err := parseTrackerResponse(resp.Body)
	if err != nil {
		log.Fatalf("failed to parse tracker response: %v", err)
	}

	fmt.Printf("Interval: %d seconds\n", interval)
	fmt.Printf("Found %d peers:\n", len(peers))

	var wg sync.WaitGroup

	for _, peer := range peers {
		wg.Add(1)
		go func(peer Peer) {
			defer wg.Done()
			conn, err := peer.ConnectAndHandshake(tf.InfoHash, [20]byte(token))
			if err == nil {
				peer.StartPeerSession(conn)
			} else {
			}
		}(peer)

		fmt.Printf(" - %s:%d\n", peer.IP.String(), peer.Port)
	}

	wg.Wait()
}
