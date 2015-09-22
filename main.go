package main

import (
	"bytes"
	"flag"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"fmt"

	"os"

	"encoding/gob"

	"github.com/peterbourgon/raft"
)

var (
	s         *raft.Server
	httpAddr  = flag.String("http", ":8080", "HTTP service address (e.g., ':8080')")
	id        = flag.Int64("id", 1, "Raft server ID. Must be > 0")
	peerAddrs = flag.String("peers", "http://127.0.0.1:8080", "HTTP service address (e.g., 'http://0.0.0.0:8080,http://0.0.0.0:8081,http://0.0.0.0:8082')")
)

// Helper function to parse URLs
func mustParseURL(rawurl string) *url.URL {
	u, err := url.Parse(rawurl)
	if err != nil {
		panic(err)
	}
	u.Path = ""
	return u
}

// Helper function to construct HTTP Peers
func mustNewHTTPPeer(u *url.URL) raft.Peer {
	p, err := raft.NewHTTPPeer(u)
	if err != nil {
		panic(err)
	}
	return p
}
func mustNewPeers() (peers []raft.Peer) {
	for _, s := range strings.Split(*peerAddrs, ",") {
		peers = append(peers, mustNewHTTPPeer(mustParseURL(s)))
	}
	return
}

func main() {
	flag.Parse()
	start()
}
func start() {
	rand.Seed(42)
	// Construct the server
	s = raft.NewServer(uint64(*id), &bytes.Buffer{}, apply)

	// Expose the server using a HTTP transport
	raft.HTTPTransport(http.DefaultServeMux, s)
	http.HandleFunc("/response", response)
	serve := func() {
		err := http.ListenAndServe(*httpAddr, nil)
		if err != nil {
			panic(err)
		} else {
			log.Println("contd")
		}

	}
	go serve()

	time.Sleep(time.Second * 5)

	peers := mustNewPeers()

	// Set the initial server configuration
	s.SetConfiguration(peers...)

	// Start the server
	s.Start()

	fmt.Println("READY!")

	dn, _ := os.Open("/dev/null")
	log.SetOutput(dn)

	for {
		cmd := make(chan []byte)

		var bytz []byte
		buf := bytes.NewBuffer(bytz)
		enc := gob.NewEncoder(buf)
		err := enc.Encode(parseQuery())
		if err != nil {
			panic(err)
		}

		if err := s.Command(buf.Bytes(), cmd); err != nil {
			panic(err)
		}
		fmt.Println(string(<-cmd))
	}
}

func apply(mid uint64, msg []byte) []byte {
	var q Query

	buf := bytes.NewBuffer(msg)
	dec := gob.NewDecoder(buf)

	err := dec.Decode(&q)
	if err != nil {
		return []byte(err.Error())
	}

	if err = q.Do(); err != nil {
		q.Msg = err.Error()
	}
	q.Id = mid
	q.ResponderID = *id

	buf.Reset()
	enc := gob.NewEncoder(buf)

	if err := enc.Encode(q); err != nil {
		return []byte(err.Error())
	}

	if resp, err := http.Post(q.Sender, "", buf); err != nil {
		return []byte(err.Error())
	} else {
		if resp.StatusCode == 500 {
			return []byte("Failed to report back")
		}
	}

	return []byte("Command sent succesfully!")

}
