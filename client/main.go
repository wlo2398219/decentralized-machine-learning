package main

import (
	"flag"
	"fmt"
	"github.com/dedis/protobuf"
	"net"
	"strconv"
	"strings"
)

// For CLI
var (
	UIPORT   = flag.Int("UIPort", 8080, "port for the UI client (default \"8080\")")
	BUDGET   = flag.Uint64("budget", 2, "budget")
	MSG      = flag.String("msg", "", "message to be sent")
	DEST     = flag.String("Dest", "", "destination for the private message")
	FILE     = flag.String("file", "", "file to be indexed by the gossiper")
	REQUEST  = flag.String("request", "", "request a chuck or metafile of this hash")
	KEYWORDS = flag.String("keywords", "", "key1,key2,key3...")
	TRAIN    = flag.Bool("train", false, "training signal")
)

// A Peerster ​simple message
type SimpleMessage struct {
	OriginalName  string // Original sender’s name
	RelayPeerAddr string // Relay Peer’s address, in the form ​ip:port
	Contents      string // The text of the message
}

type RumorMessage struct {
	Origin string
	ID     uint32
	Text   string
}
type PeerStatus struct {
	Identifier string
	NextID     uint32
}

type StatusPacket struct {
	Want []PeerStatus // vector clock
}

type PrivateMessage struct {
	Origin      string
	ID          uint32
	Text        string
	Destination string
	HopLimit    uint32
}

type DataRequest struct {
	Origin      string
	Destination string
	HopLimit    uint32
	HashValue   []byte
}

type DataReply struct {
	Origin      string
	Destination string
	HopLimit    uint32
	HashValue   []byte
	Data        []byte
}

// the ​ONLY​ type of packets sent to other peers
type SearchRequest struct {
	Origin   string
	Budget   uint64
	Keywords []string
}

type SearchReply struct {
	Origin      string
	Destination string
	HopLimit    uint32
	Results     []*SearchResult
}

type SearchResult struct {
	FileName     string
	MetafileHash []byte
	ChunkMap     []uint64
}

// The packet eventually transmitted between nodes
type GossipPacket struct {
	Simple        *SimpleMessage
	Rumor         *RumorMessage
	Status        *StatusPacket
	Private       *PrivateMessage
	DataRequest   *DataRequest
	DataReply     *DataReply
	SearchRequest *SearchRequest
	SearchReply   *SearchReply
}

type Gossiper struct {
	address *net.UDPAddr
	conn    *net.UDPConn
	Name    string
}

func NewGossiper(address, name string) *Gossiper {
	// udpAddr, err := net.ResolveUDPAddr("udp4", address)
	// udpConn, err := net.ListenUDP("udp4", udpAddr)
	udpAddr, _ := net.ResolveUDPAddr("udp4", address)
	udpConn, _ := net.ListenUDP("udp4", udpAddr)

	return &Gossiper{
		address: udpAddr,
		conn:    udpConn,
		Name:    name,
	}
}

type Message struct {
	Text string
}

func main() {

	var packetToSend GossipPacket

	flag.Parse()

	if *TRAIN && *FILE != "" {
		simpleMessage := &SimpleMessage{Contents: *FILE, OriginalName: "TRAIN"}
		packetToSend = GossipPacket{Simple: simpleMessage}
	} else if *DEST == "" && *MSG != "" { // simple message
		simpleMessage := &SimpleMessage{Contents: *MSG, OriginalName: "RUMOR"}
		packetToSend = GossipPacket{Simple: simpleMessage}
	} else if *DEST != "" && *MSG != "" { // private message
		privateMessage := &PrivateMessage{Destination: *DEST, ID: 0, Text: *MSG, HopLimit: 10}
		packetToSend = GossipPacket{Private: privateMessage}
	} else if *FILE != "" && *REQUEST == "" { // scan file
		simpleMessage := &SimpleMessage{Contents: *FILE, OriginalName: "FILE"}
		packetToSend = GossipPacket{Simple: simpleMessage}
	} else if *FILE != "" && *REQUEST != "" { // download file
		fmt.Println("FILE DOWNLOAD REQUEST:", *FILE, "FROM", *DEST)
		privateMessage := &PrivateMessage{Destination: *DEST, Text: *REQUEST, HopLimit: 10}
		simpleMessage := &SimpleMessage{Contents: *FILE, OriginalName: "FILE_DOWNLOAD"}
		packetToSend = GossipPacket{Simple: simpleMessage, Private: privateMessage}
	} else if *KEYWORDS != ""{
		words := strings.Split(*KEYWORDS, ",")
		fileSearchMessage := &SearchRequest{Budget: *BUDGET, Keywords: words}
		packetToSend = GossipPacket{SearchRequest: fileSearchMessage}
	} else {
		return
	}

	conn, _ := net.Dial("udp4", "127.0.0.1:"+strconv.Itoa(*UIPORT))
	packetBytes, _ := protobuf.Encode(&packetToSend)
	conn.Write(packetBytes)

}
