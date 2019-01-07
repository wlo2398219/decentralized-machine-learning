package main

import (
	"fmt"
	"math/rand"
	"net"
	"time"
	"github.com/dedis/protobuf"
	
)

// func rumorMongering(conn *net.UDPConn, sender, mongerIP string, ch chan bool, msgBytes []byte) {
func rumorMongering(conn *net.UDPConn, sender, mongerIP string, ch chan *StatusPacket, msgBytes []byte) {

	// fmt.Println("MONGERING with", mongerIP)
	var msg = GossipPacket{}

	protobuf.Decode(msgBytes, &msg)
	err := sendPacketToAddr(conn, msg, mongerIP)

	// dst, _ := net.ResolveUDPAddr("udp4", mongerIP)
	// _, err := conn.WriteToUDP(msgBytes, dst)

	if err != nil {
		fmt.Println("ERROR in Mongering~~~", err, len(msgBytes))
	}

	if sender != "" {
		// send back status packet to the one who sends rumor to you

		err = sendPacketToAddr(conn, GossipPacket{Status: &status}, sender)
		if err != nil {
			fmt.Println("ERROR in Mongering!!!", err)
		} else {
			// fmt.Println("NO ERROR in Mongering!!!")
		}
	}

	var timer = time.NewTimer(time.Second)

	select {
	case <-timer.C:
		if continueSending() && peer_list.Len() != 1 {
			nextOne := randomPeer(mongerIP)
			// fmt.Println("FLIPPED COIN sending rumor to", nextOne)
			go rumorMongering(conn, mongerIP, nextOne, recv_channels[nextOne], msgBytes)
		} else {
		}
	case s := <-ch:
		go compareStatusAndSend(conn, s, mongerIP, msgBytes)
	}

	return

}

// return the elements that slice1 doesn't have
func statusDifference(slice1, slice2 []PeerStatus) []PeerStatus {
	var diff = []PeerStatus{}
	var sliceMap = map[string]uint32{}

	// create a map
	for _, s := range slice1 {
		sliceMap[s.Identifier] = s.NextID
	}

	for _, s := range slice2 {
		_, exist := sliceMap[s.Identifier]

		if !exist && s.NextID == 1 {
			continue
		} else if !exist {
			diff = append(diff, PeerStatus{Identifier: s.Identifier, NextID: 1})
		} else if exist && sliceMap[s.Identifier] < s.NextID {
			diff = append(diff, PeerStatus{Identifier: s.Identifier, NextID: sliceMap[s.Identifier]})
		}
	}

	return diff
}

func compareStatusAndSend(conn *net.UDPConn, recv_status *StatusPacket, dst_addr string, msgBytes []byte) {
	var (
		text string
		diff []PeerStatus
	)

	diff = statusDifference(recv_status.Want, status.Want)

	if len(diff) != 0 {
		s := diff[0]
		text = nodeName_msgs[s.Identifier][s.NextID-1]
		err := sendPacketToAddr(conn, GossipPacket{Rumor: &RumorMessage{Origin: s.Identifier, ID: s.NextID, Text: text}}, dst_addr)

		// fmt.Println("SEND PACKET", s.Identifier, status.Want[status_index[s.Identifier]].NextID, s.NextID)
		if err != nil {
			fmt.Println("ERROR in compareStatusAndSend", err)
		}

		return
	}

	diff = statusDifference(status.Want, recv_status.Want)

	if len(diff) != 0 {
		// fmt.Println("SEND NEW MSG FROM newMsg")
		err := sendPacketToAddr(conn, GossipPacket{Status: &status}, dst_addr)
		if err != nil {
			fmt.Println("ERROR in compareStatusAndSend", err)
		}
	}

	// fmt.Println("IN SYNC WITH " + dst_addr)
	if msgBytes != nil && continueSending() && peer_list.Len() != 1 {
		nextOne := randomPeer(dst_addr)
		// fmt.Println("FLIPPED COIN sending rumor to", nextOne)
		go rumorMongering(conn, dst_addr, nextOne, recv_channels[nextOne], msgBytes)
	}

	return
}

func antiEntropy(conn *net.UDPConn) {
	ticker := time.NewTicker(time.Second)

	for _ = range ticker.C {
		if peer_list.Len() > 0 {
			dst := randomPeer("")
			err := sendPacketToAddr(conn, GossipPacket{Status: &status}, dst)

			if err != nil {
				fmt.Println(dst)
				fmt.Println("ERROR in  ANTIENTROPY")
			} else {
				// fmt.Println("DONE IN antiEntropy")
			}

		}

	}

}

func antiEntropyReceiver(conn *net.UDPConn, ch chan *StatusPacket, sender string) {

	for {
		s := <-ch
		go compareStatusAndSend(conn, s, sender, nil)
	}
}

func updateStatus(orgSender string, msgID uint32) { // append new status

	_, exist := status_index[orgSender]

	if !exist {
		nodeName_msgs[orgSender] = []string{}
		status.Want = append(status.Want, PeerStatus{Identifier: orgSender, NextID: 1})
		status_index[orgSender] = len(status.Want) - 1
	}

}

func continueSending() bool {
	if rand.Int()%2 == 1 {
		return true
	} else {
		return false
	}
}
