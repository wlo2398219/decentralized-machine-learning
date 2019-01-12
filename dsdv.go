package main

import (
	// "fmt"
	"time"
)

type NextHopPair struct {
	NextHop string
	SeqNum  uint32
}

var nextHopTable = map[string]*NextHopPair{}

func updateNextHop(orgSender, relaySender string, ID uint32) {

	_, exist := nextHopTable[orgSender]

	if orgSender == *name {
		return
	}

	if !exist {
		downloadProcessTable[orgSender] = make(chan *DataReply)
		nextHopTable[orgSender] = &NextHopPair{NextHop: relaySender, SeqNum: ID}
		// fmt.Println("DSDV", orgSender, relaySender)
	} else if nextHopTable[orgSender].SeqNum < ID && nextHopTable[orgSender].NextHop != relaySender {
		// fmt.Println("DSDV - ", nextHopTable[orgSender].NextHop, ",", nextHopTable[orgSender].SeqNum, ",", relaySender, ",", ID)
		nextHopTable[orgSender].NextHop = relaySender
		nextHopTable[orgSender].SeqNum = ID
		// fmt.Println("DSDV", orgSender, relaySender)
	}
}

func routeRumor(ch chan *GossipPacket, duration int) {

	ticker := time.NewTicker(time.Second * time.Duration(duration))

	packet := GossipPacket{Simple: &SimpleMessage{OriginalName: "RUMOR", Contents: ""}}
	ch <- &packet

	for {
		<-ticker.C
		packet = GossipPacket{Simple: &SimpleMessage{OriginalName: "RUMOR", Contents: ""}}
		ch <- &packet
	}

}
