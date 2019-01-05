package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/dedis/protobuf"
	"github.com/gorilla/handlers"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Global variable
var (
	UIPort     = flag.Int("UIPort", 8080, "port for the UI client (default \"8080\")")
	rtimer     = flag.Int("rtimer", 0, "route rumors sending period in seconds, 0 to disable sending of route rumors (default 0)")
	gossipAddr = flag.String("gossipAddr", "127.0.0.1:5000", "ip:port for the gossiper (default \"127.0.0.1:5000\")")
	name       = flag.String("name", "281544", "name of the gossiper")
	peers      = flag.String("peers", "", "comma seperated list of peers of the form ip:port")
	simple     = flag.Bool("simple", false, "run gossiper in simple broadcast mode")
	gui        = flag.Bool("gui", false, "run gossiper in gui mode")

	status StatusPacket

	status_index  = map[string]int{}
	nodeName_msgs = map[string][]string{}
	peer_index    = map[string]int{}

	recv_channels        = map[string]chan *StatusPacket{}
	downloadProcessTable = map[string]chan *DataReply{}

	ordered_msgs = ConcurrentSliceString{items: []string{}}
	peer_list    = &ConcurrentSliceString{items: []string{}}

	gossiper_peer *Gossiper

	MAX_PACKET_SIZE = 102400
)

func main() {
	// parse the flags
	flag.Parse()
	rand.Seed(time.Now().UTC().UnixNano())

	status.Want = append(status.Want, PeerStatus{Identifier: *name, NextID: 1})
	status_index[*name] = 0
	nodeName_msgs[*name] = []string{}
	gossiper_peer = NewGossiper(*gossipAddr, *name)
	UIAddr := "127.0.0.1:" + strconv.Itoa(*UIPort)
	chClientPacket := make(chan *GossipPacket)

	var (
		wg sync.WaitGroup

		// for HW2
		fileData = map[string]*mFile{}

		// for HW3
	)
	wg.Add(1)

	if *peers != "" {

		for ind, peer := range strings.Split(*peers, ",") {
			peer_list.Append(peer)
			peer_index[peer] = ind
			recv_channels[peer] = make(chan *StatusPacket)
			go antiEntropyReceiver(gossiper_peer.conn, recv_channels[peer], peer)
		}

	}

	if *rtimer != 0 {
		fmt.Println("OPEN ROUTE RUMOR")
		// go routeRumor(UIAddr, *rtimer)
		go routeRumor(chClientPacket, *rtimer)

	}

	// UI part
	// go handleClient(UIAddr, gossiper_peer)

	go recvPacket(UIAddr, chClientPacket)
	go handleClient(chClientPacket, gossiper_peer, fileData)

	// gossiper part
	go handleGossiper(gossiper_peer, fileData)

	if !*simple {
		// anti-entropy
		go antiEntropy(gossiper_peer.conn)
	}

	// go PoW(gossiper_peer.conn, peer_list)

	// GUI part
	if *gui {

		time.Sleep(time.Second)

		r := http.NewServeMux()
		r.Handle("/", http.FileServer(http.Dir(".")))
		r.HandleFunc("/msg/", messageHandler(chClientPacket))
		r.HandleFunc("/node/", nodeHandler)
		r.HandleFunc("/id/", idHandler)
		r.HandleFunc("/file/", fileHandler(chClientPacket))
		r.HandleFunc("/p2pDownload/", p2pDownloadHandler(chClientPacket, *name, gossiper_peer.conn, peer_list))

		http.ListenAndServe(":8080", handlers.CompressHandler(r))

	}

	wg.Wait()

}

func handleGossiper(gossiper *Gossiper, fileData map[string]*mFile) {
	var (
		// msg         GossipPacket
		relaySender string
		orgSender   string
		addr        *net.UDPAddr
		// chSearchRequest *SearchRequest
		// SearchRequestMap = ConcurrentMap{items: map[string]bool{}}

	)

	for {

		packetBytes := make([]byte, MAX_PACKET_SIZE)
		msg := GossipPacket{}
		_, addr, _ = gossiper.conn.ReadFromUDP(packetBytes)
		relaySender = addr.String()
		protobuf.Decode(packetBytes, &msg)

		// append new peer
		if !peer_contained(relaySender) {
			peer_list.Append(relaySender)
			peer_index[relaySender] = peer_list.Len() - 1

			recv_channels[relaySender] = make(chan *StatusPacket)
			go antiEntropyReceiver(gossiper.conn, recv_channels[relaySender], relaySender)

		}

		if *simple && msg.Simple != nil {

			//  SIMPLE MESSAGE origin <original_sender_name> from <relay_addr> contents <msg_text>
			fmt.Println("SIMPLE MESSAGE origin", msg.Simple.OriginalName,
				"from", msg.Simple.RelayPeerAddr,
				"contents", msg.Simple.Contents)

			// if receiving packet from itself, stop sending
			if msg.Simple.OriginalName == *name {
				continue
			}

			if peer_list.Len() != 0 {

				printPeerInfo()
				msg.Simple.RelayPeerAddr = *gossipAddr
				packetBytes, _ = protobuf.Encode(&msg)
				sendToPeers(gossiper.conn, relaySender, packetBytes)
			}
		} else if msg.Rumor != nil { // RUMOR PACKET
			orgSender = msg.Rumor.Origin
			if msg.Rumor.Text != "" { // CHAT RUMOR
				fmt.Printf("RUMOR origin %s from %s ID %s contents %s\n",
					orgSender, relaySender, strconv.Itoa(int(msg.Rumor.ID)), msg.Rumor.Text)
				printPeerInfo()
			}

			updateStatus(orgSender, msg.Rumor.ID)
			updateIDlist(orgSender)

			// in-order new gossip message
			if status.Want[status_index[orgSender]].NextID == msg.Rumor.ID {
				updateNextHop(orgSender, relaySender, msg.Rumor.ID)
				status.Want[status_index[orgSender]].NextID++
				nodeName_msgs[orgSender] = append(nodeName_msgs[orgSender], msg.Rumor.Text)

				if msg.Rumor.Text != "" {
					ordered_msgs.Append(orgSender + ":" + msg.Rumor.Text)
				}

				if peer_list.Len() != 1 {

					mongerIP := randomPeer(relaySender)
					go rumorMongering(gossiper.conn, relaySender, mongerIP, recv_channels[mongerIP], packetBytes)
				}
			}

		} else if msg.Status != nil { // STATUS PACKET
			fmt.Print("STATUS from " + relaySender)
			for _, s := range msg.Status.Want {
				fmt.Print(" peer " + s.Identifier + " nextID " + strconv.Itoa(int(s.NextID)))
			}
			fmt.Println()
			printPeerInfo()
			recv_channels[relaySender] <- msg.Status

		} else if msg.Private != nil {
			if msg.Private.Destination == *name {
				// ordered_msgs = append(ordered_msgs, msg.Private.Origin+" (PRIVATE):"+msg.Private.Text)
				ordered_msgs.Append(msg.Private.Origin + " (PRIVATE):" + msg.Private.Text)

				fmt.Println("PRIVATE origin", msg.Private.Origin, "hop-limit", msg.Private.HopLimit, "contents", msg.Private.Text)
			} else if _, exist := nextHopTable[msg.Private.Destination]; exist {
				if msg.Private.HopLimit > 1 {
					msg.Private.HopLimit--
					_ = sendPacketToAddr(gossiper.conn, msg, nextHopTable[msg.Private.Destination].NextHop)
					fmt.Println("---- FORWARD PRIVATE MESSAGE TO", nextHopTable[msg.Private.Destination].NextHop, "DESTINATION:", msg.Private.Destination, "HOP-LIMIT", msg.Private.HopLimit, "----")
				} else {
					fmt.Println("---- HopLimit becomes 0 for packet to", msg.Private.Destination, "----")
				}
			}
		} else if msg.DataRequest != nil { // file download

			if msg.DataRequest.Destination == *name {
				go searchHashAndSend(gossiper.conn, msg.DataRequest.HashValue, *name, msg.DataRequest.Origin)
			} else if _, exist := nextHopTable[msg.DataRequest.Destination]; exist && msg.DataRequest.HopLimit > 1 {
				msg.DataRequest.HopLimit--
				_ = sendPacketToAddr(gossiper.conn, msg, nextHopTable[msg.DataRequest.Destination].NextHop)
			}

		} else if msg.DataReply != nil {

			if msg.DataReply.Destination == *name {
				handleDataReplies(msg.DataReply)
				// downloadProcessTable[msg.DataReply.Origin] <- msg.DataReply
			} else if _, exist := nextHopTable[msg.DataReply.Destination]; exist && msg.DataReply.HopLimit > 1 {
				msg.DataReply.HopLimit--
				_ = sendPacketToAddr(gossiper.conn, msg, nextHopTable[msg.DataReply.Destination].NextHop)
			}

		} else if msg.SearchRequest != nil {
			fmt.Println("--- GOT FILESEARCH IN GOSSIPER ---")
			fmt.Println("--- FILE SEARCH WITH KEYWORDS ---", msg.SearchRequest.Keywords)
			fmt.Println("--- W/ BUDGET ---", msg.SearchRequest.Budget)

			handleSearchRequest(gossiper.conn, msg.SearchRequest, fileData, *name, relaySender, peer_list)
			// sendSearchReplies(gossiper.conn, msg.SearchRequest.Keywords, fileData, *name, msg.SearchRequest.Origin)
			// distributeSearch(gossiper.conn, relaySender, *name, msg.SearchRequest.Budget, msg.SearchRequest.Keywords, peer_list)
		} else if msg.SearchReply != nil {

			if msg.SearchReply.Destination == *name {
				// fmt.Println("--- GOT SEARCHREPLY --- FROM ", msg.SearchReply.Origin)
				// for _, val := range msg.SearchReply.Results {
				// 	fmt.Println("#CHUNKS:", val.ChunkCount)
				// 	fmt.Printf("%s, %x\n", val.FileName, val.MetafileHash)
				// 	fmt.Println(val.ChunkMap)
				// }

				go handleSearchReplies(msg.SearchReply)
			} else if _, exist := nextHopTable[msg.SearchReply.Destination]; exist && msg.SearchReply.HopLimit > 1 {
				msg.SearchReply.HopLimit--
				_ = sendPacketToAddr(gossiper.conn, msg, nextHopTable[msg.SearchReply.Destination].NextHop)
			}

		} else if msg.TxPublish != nil {

			fmt.Println("===== RECEIVE TXPUBLISH =====")
			fmt.Println(msg.TxPublish.File.Name)
			fmt.Println(msg.TxPublish.HopLimit)
			handlebcFile(gossiper.conn, msg.TxPublish, peer_list, relaySender)

		} else if msg.BlockPublish != nil {

			// fmt.Println("===== RECEIVE NEW BLOCK =====")
			// fmt.Printf("%x\n",msg.BlockPublish.Block.PrevHash)
			// fmt.Printf("%x\n",msg.BlockPublish.Block.Hash())
			// fmt.Println(msg.BlockPublish.HopLimit)
			handleBlock(gossiper.conn, msg.BlockPublish, peer_list, relaySender)
		} else if msg.WeightPacket != nil {
			// fmt.Println("GET THE WEIGHT PACKET!")
			handleWeight(gossiper.conn, msg.WeightPacket)
			// sendToPeers(gossiper.conn, "", packetBytes)
		} else if msg.GradientPacket != nil {

			handleGradient(gossiper.conn, msg.GradientPacket)
		} else {
			fmt.Println("all-nil message")
		}

	}

}

func recvPacket(address string, ch chan *GossipPacket) {
	udpAddr, _ := net.ResolveUDPAddr("udp4", address)
	udpConn, _ := net.ListenUDP("udp4", udpAddr)

	for {
		msg := GossipPacket{}
		packetBytes := make([]byte, MAX_PACKET_SIZE)
		udpConn.ReadFromUDP(packetBytes)
		protobuf.Decode(packetBytes, &msg)
		ch <- &msg
	}
}

// func handleClient(address string, gossiper *Gossiper) {
// func handleClient(ch chan *GossipPacket, gossiper *Gossiper, shareFiles map[string]bool, fileChunkIndex map[string]uint64) {
func handleClient(ch chan *GossipPacket, gossiper *Gossiper, fileData map[string]*mFile) {

	var (
		msg *GossipPacket
	)
	for {

		packetBytes := make([]byte, MAX_PACKET_SIZE)
		msg = <-ch

		if msg.Simple != nil && msg.Simple.Contents != "" {
			fmt.Println("CLIENT MESSAGE", msg.Simple.Contents)
		} else if msg.Private == nil && msg.Simple == nil && msg.SearchRequest == nil {
			fmt.Println("ERROR HAPPENS IN client side")
		}

		if *simple { // SIMPLE MODE
			fmt.Println("HELLO")
			if peer_list.Len() != 0 {
				fmt.Println("HELLO AGAIN")
				msg.Simple.OriginalName = *name
				msg.Simple.RelayPeerAddr = *gossipAddr
				packetBytes, _ = protobuf.Encode(msg)

				sendToPeers(gossiper.conn, "", packetBytes)
			}

		} else if msg.Simple != nil {
			switch msg.Simple.OriginalName {
			case "RUMOR":
				if msg.Simple.Contents != "" {
					// ordered_msgs = append(ordered_msgs, *name+":"+msg.Simple.Contents)
					ordered_msgs.Append(*name + ":" + msg.Simple.Contents)
				}

				nodeName_msgs[*name] = append(nodeName_msgs[*name], msg.Simple.Contents)
				status.Want[0].NextID++

				// start rumor mongering here
				msg.Rumor = &RumorMessage{Origin: *name, ID: status.Want[0].NextID - 1, Text: msg.Simple.Contents}
				msg.Simple = nil
				packetBytes, _ = protobuf.Encode(msg)

				if peer_list.Len() > 0 {

					mongerIP := randomPeer("")
					go rumorMongering(gossiper.conn, "", mongerIP, recv_channels[mongerIP], packetBytes)
				}

			case "FILE":
				f := GetFile(msg.Simple.Contents, fileData)
				fmt.Println("UPLOAD FILE " + msg.Simple.Contents + " SUCCEEDS")

				go publishFile(gossiper.conn, f, peer_list)
				// PrintFileInfo(fileData[msg.Simple.Contents])

			case "FILE_DOWNLOAD":
				fmt.Println("----- FILE DOWNLOAD REQUEST -----")
				data, _ := hex.DecodeString(msg.Private.Text)
				fileName := msg.Simple.Contents
				if msg.Private.Destination == "" && searchFileTable.Exist(fileName) {
					// fmt.Println("P2P FILE_DOWNLOAD")
					dstArr := make([]string, searchFileTable.Get(fileName).(sFile).ChunkCount)

					for ind, dst := range searchFileTable.Get(fileName).(sFile).AvailableChunk {
						dstArr[ind-1] = dst
					}

					go DownloadFile(gossiper.conn, fileName, *name, dstArr, data, fileData)

				} else if _, exist := nextHopTable[msg.Private.Destination]; exist {
					// fmt.Println("FILE_DOWNLOAD")
					go DownloadFile(gossiper.conn, fileName, *name, msg.Private.Destination, data, fileData)
					// go DownloadFile(gossiper.conn, msg.Simple.Contents, *name, msg.Private.Destination, data, downloadProcessTable[msg.Private.Destination])
				} else {
					fmt.Println("DON'T KNOW WHERE THE DESTINATION IS")
				}
			case "TRAIN":
				fmt.Println("---- TRAINING REQUEST ----", msg.Simple.Contents)
				newTraining(gossiper.conn, "mnist")  //  dataset
				// newTraining(gossiper.conn, "uci_cbm_dataset.txt")  //  dataset
				
			}
		} else if msg.Private != nil { // PRIVATE MESSAGE

			if _, exist := nextHopTable[msg.Private.Destination]; exist {
				fmt.Println("---- SEND PRIVATE MESSAGE TO", msg.Private.Destination, "via", nextHopTable[msg.Private.Destination].NextHop, "----")
				msg.Private.Origin = *name
				_ = sendPacketToAddr(gossiper.conn, *msg, nextHopTable[msg.Private.Destination].NextHop)

			} else {
				fmt.Println("DON'T KNOW WHERE", msg.Private.Destination, "IS")
			}
		} else if msg.SearchRequest != nil {
			// fmt.Println("--- FILE SEARCH WITH KEYWORDS ---", msg.SearchRequest.Keywords)
			// fmt.Println("--- W/ BUDGET ---", msg.SearchRequest.Budget)
			// distributeSearch(gossiper.conn, "", *name, msg.SearchRequest.Budget+1, msg.SearchRequest.Keywords, peer_list)

			go newSearchProcess(gossiper.conn, *name, msg.SearchRequest.Budget, msg.SearchRequest.Keywords, peer_list)
			// distributeSearch(gossiper.conn, "", *name, msg.SearchRequest.Budget+1, msg.SearchRequest.Keywords, peer_list)

		} else {
			fmt.Println("UNKNOWN DATA FROM CLIENT SIDE")
		}

	}
}

func sendPacketToAddr(conn *net.UDPConn, gossipPacket GossipPacket, dst_addr string) error {
	packetBytes := make([]byte, MAX_PACKET_SIZE)
	msg := &gossipPacket
	packetBytes, err1 := protobuf.Encode(msg)
	dst, _ := net.ResolveUDPAddr("udp4", dst_addr)
	_, err := conn.WriteToUDP(packetBytes, dst)

	if err1 != nil {
		fmt.Println("ERROR!!!!!!! ")
		fmt.Println(err1)
	}
	// if gossipPacket.WeightPacket != nil {
		// fmt.Println("==== SEND WEIGHTPACKET TO", dst_addr, "====")
		// fmt.Println(gossipPacket.WeightPacket.Weight)
		// fmt.Println(packetBytes)
	// }

	if err != nil {
		log.Fatal(err)
	}

	return err
}

func randomPeer(excludedPeer string) string {
	if excludedPeer == "" {
		return peer_list.Get(rand.Intn(peer_list.Len()))

	} else {
		ind := peer_index[excludedPeer]
		randNum := (ind + rand.Intn(peer_list.Len()-1) + 1) % peer_list.Len()

		return peer_list.Get(randNum)

	}
}

func NewGossiper(address, name string) *Gossiper {
	udpAddr, _ := net.ResolveUDPAddr("udp4", address)
	udpConn, _ := net.ListenUDP("udp4", udpAddr)

	return &Gossiper{
		address: udpAddr,
		conn:    udpConn,
		Name:    name,
	}
}

func sendToPeers(conn *net.UDPConn, sender string, packetBytes []byte) {
	for peer := range peer_list.Iter() {
		if peer != sender {
			fmt.Println("SEND TO ", peer)
			fmt.Println(packetBytes)
			dst, _ := net.ResolveUDPAddr("udp4", peer)
			_, err := conn.WriteToUDP(packetBytes, dst)

			if err != nil {
				fmt.Println("ERROR IN sendToPeers:", err)
			}

		}
	}
}

// func peer_contained(peers []string, peer string) bool {
func peer_contained(peer string) bool {

	if _, exist := peer_index[peer]; !exist && peer != *gossipAddr {
		return false
	} else {
		return true
	}

}

func printPeerInfo() {
	fmt.Print("PEERS ", peer_list.Get(0))

	for j := 1; j < peer_list.Len(); j++ {
		fmt.Print("," + peer_list.Get(j))
	}

	fmt.Println()
}
