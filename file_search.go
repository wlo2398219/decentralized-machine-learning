package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"net"
	"regexp"
	"strings"
	"time"
)

var (
	duplicateCheckMap = ConcurrentMap{items: map[interface{}]interface{}{}}
	searchProcessMap  = ConcurrentMap{items: map[interface{}]interface{}{}}
	searchFileTable   = ConcurrentMap{items: map[interface{}]interface{}{}}
	keywordsTable     = ConcurrentMap{items: map[interface{}]interface{}{}}
)

type sChunck struct {
	Index       uint64
	Destination string
}

type sFile struct {
	FileName       string
	ChunkCount     uint64
	AvailableChunk map[uint64]string
	MetafileHash   []byte
}

func isDuplicate(key interface{}, m ConcurrentMap) bool {
	if exist := m.Exist(key); exist {
		return true
	} else {
		return false
	}
}

func registerNewSearchRequest(key string, m ConcurrentMap) {

	m.Set(key, true)
	timer := time.NewTimer(time.Millisecond * 500)

	select {
	case <-timer.C:
		m.Delete(key)
		fmt.Println("===== DELETE =====")
	}

}

// handle search request from other peers
func handleSearchRequest(conn *net.UDPConn, data *SearchRequest, fileData map[string]*mFile, myName, relaySender string, peers *ConcurrentSliceString) {

	key := data.Origin + "," + strings.Join(data.Keywords, ",")
	if isDuplicate(key, duplicateCheckMap) {
		fmt.Println("===== DUPLICATE =====")
		return
	} else {
		// ignore duplicate within 0.5 second
		go registerNewSearchRequest(key, duplicateCheckMap)
	}

	// search locally and send reply
	sendSearchReplies(conn, data.Keywords, fileData, myName, data.Origin)
	// distribute the request to other peers
	distributeSearch(conn, relaySender, data.Origin, data.Budget, data.Keywords, peers)

}

func sendSearchReplies(conn *net.UDPConn, keywords []string, fileData map[string]*mFile, org, dst string) {
	var (
		pattern = ".*("
		results = []*SearchResult{}
		// chunkMap []uint64
	)

	// create regexp
	pattern += keywords[0]
	for _, word := range keywords[1:] {
		pattern += "|" + word
	}
	pattern += ").*"

	re, err := regexp.Compile(pattern)
	if err != nil {
		log.Fatal(err)
	}

	for name := range fileData {

		if re.MatchString(name) {

			results = append(results, &SearchResult{FileName: name, MetafileHash: (fileData[name]).MetaHash, ChunkMap: fileData[name].AvailableIndex, ChunkCount: (fileData[name]).ChunkCount})
		}
		fmt.Println(name)
	}

	if len(results) != 0 {
		// fmt.Println(results[0])
		// fmt.Println("===== SEND SEARCHREPLY FROM", org, "TO", dst, "=====")
		searchReply := SearchReply{Origin: org, Destination: dst, HopLimit: 10, Results: results}
		sendPacketToAddr(conn, GossipPacket{SearchReply: &searchReply}, nextHopTable[dst].NextHop)

	}

}

// func distributeSearch(conn *net.UDPConn, orgAddr, senderName string, budget uint64, words, peers []string) {
func distributeSearch(conn *net.UDPConn, relaySender, org string, budget uint64, words []string, peers *ConcurrentSliceString) {

	var (
		// budgetArr         = make([]uint64, len(peers))
		budgetArr = make([]uint64, peers.Len())
		// tmp               = uint64(len(peers))
		tmp = uint64(peers.Len())

		count             uint64
		fileSearchMessage *SearchRequest
		packetToSend      GossipPacket
	)

	budget--

	if relaySender != "" {
		tmp--
	}

	if budget == 0 || tmp == 0 {
		return
	}

	quotient, remainder := budget/tmp, budget%tmp

	// assign budgets
	count = 0

	for _, ind := range rand.Perm(peers.Len()) {
		if peers.Get(ind) != relaySender && count < remainder {
			budgetArr[ind] = quotient + 1
			count++

		} else if peers.Get(ind) != relaySender {
			budgetArr[ind] = quotient
		}
	}

	for ind, b := range budgetArr {
		if b != 0 {
			fileSearchMessage = &SearchRequest{Budget: b, Keywords: words, Origin: org}
			packetToSend = GossipPacket{SearchRequest: fileSearchMessage}
			sendPacketToAddr(conn, packetToSend, peers.Get(ind))
		}
	}

}

// handle new search request from client
func newSearchProcess(conn *net.UDPConn, org string, budget uint64, keywords []string, peers *ConcurrentSliceString) (string, string) {

	fileTraceTable := map[string]sFile{}

	count := 0
	ch := make(chan *SearchReply)
	double := budget == 2
	completeFiles := []sFile{}

	pattern := ".*("
	pattern += keywords[0]
	for _, word := range keywords[1:] {
		pattern += "|" + word
	}
	pattern += ").*"
	re, err := regexp.Compile(pattern)
	if err != nil {
		log.Fatal(err)
	}

	// if isDuplicate(key, searchProcessMap) {
	if isDuplicate(re, searchProcessMap) {
		fmt.Println("===== DUPLACE SEARCH ====")
		return "", ""
	} else {
		// create new channel to get SearchReply

		keywordsTable.Set(re, re)
		searchProcessMap.Set(re, ch)

	}

	distributeSearch(conn, "", org, budget+1, keywords, peers)

	ticker := time.NewTicker(time.Second)

	for {
		select {
		case <-ticker.C:
			if double && budget < 32 {
				budget *= 2
				fmt.Println("======= SEND WITH BUDGET", budget, "=======")
				distributeSearch(conn, "", org, budget+1, keywords, peers)
			} else {
				fmt.Println("======= EXCEED BUDGET =======")
				close(ch)
				searchProcessMap.Delete(re)
				return "", ""
			}
		case packet := <-ch:
			// fmt.Println("======= PROCESS GETTING NEW SEARCHREPLY PACKET =======")
			for _, data := range packet.Results {

				if _, exist := fileTraceTable[data.FileName]; !exist { // new file during the search
					fileTraceTable[data.FileName] = sFile{FileName: data.FileName, ChunkCount: data.ChunkCount, MetafileHash: data.MetafileHash, AvailableChunk: map[uint64]string{}}
				}

				fmt.Println("FOUND match", data.FileName, "at", packet.Origin)
				fmt.Printf("metafile=%s ", hex.EncodeToString(data.MetafileHash))
				for ind, chunkIndex := range data.ChunkMap {
					if ind == 0 {
						fmt.Printf("chunks=%d", chunkIndex)
					} else {
						fmt.Printf(",%d", chunkIndex)
					}
					if _, exist := fileTraceTable[data.FileName].AvailableChunk[chunkIndex]; !exist {
						fileTraceTable[data.FileName].AvailableChunk[chunkIndex] = packet.Origin
					}
				}
				fmt.Println()

				if uint64(len(fileTraceTable[data.FileName].AvailableChunk)) == fileTraceTable[data.FileName].ChunkCount {
					count++
					completeFiles = append(completeFiles, fileTraceTable[data.FileName])

					if count == 2 {
						fmt.Println("SEARCH FINISHED")
						break
					}
				}
			}
		}
		if count == 2 {
			break
		}
	}

	nameArr := []string{}

	for _, file := range completeFiles {
		searchFileTable.Set(file.FileName, file)
		// fmt.Println("~~~~",file.FileName)
		nameArr = append(nameArr, file.FileName)
	}
	// fmt.Println("@_@_@_@_@_@_@_@_",nameArr)
	// fmt.Println("@_@_@_@_@_@_@_@_:",nameArr[0],",",nameArr[1])

	// fmt.Println("======= CLOSE PROCESS =======")
	close(ch)
	// searchProcessMap.Delete(key)
	// keywordsTable.Delete(key)
	searchProcessMap.Delete(re)
	keywordsTable.Delete(re)
	// fmt.Println(nameArr)
	// fmt.Println("@_@_@_@_@_@_@_@_:",nameArr[0],",",nameArr[1])
	return nameArr[0], nameArr[1]
}

func handleSearchReplies(packet *SearchReply) {
	fileNames := []string{}

	for _, data := range packet.Results {
		fileNames = append(fileNames, data.FileName)
	}

	for re := range keywordsTable.Iter() {
		pass := true
		for _, name := range fileNames {
			pass = pass && re.(*regexp.Regexp).MatchString(name)
		}

		if pass {
			searchProcessMap.Get(re).(chan *SearchReply) <- packet
		}
	}
}
