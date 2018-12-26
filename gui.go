package main

import (
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"strings"
)

var id_list []string

func messageHandler(ch chan *GossipPacket) func(w http.ResponseWriter, req *http.Request) {

	return func(w http.ResponseWriter, req *http.Request) {

		switch req.Method {
		case "POST":

			text := req.FormValue("text")

			if req.FormValue("dest") == "" {
				simplemessage := &SimpleMessage{OriginalName: "RUMOR", RelayPeerAddr: "", Contents: text}
				ch <- &GossipPacket{Simple: simplemessage}
			} else {
				privateMessage := &PrivateMessage{Destination: req.FormValue("dest"), ID: 0, Text: text, HopLimit: 10}
				ch <- &GossipPacket{Private: privateMessage}
			}

		case "GET":
			fmt.Fprintf(w, strings.Join(ordered_msgs.GetSlice(), "<br/>"))

		case "default":
			fmt.Println("NO IDEA")
		}
	}
}

func fileHandler(ch chan *GossipPacket) func(w http.ResponseWriter, req *http.Request) {

	return func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case "POST":
			fmt.Println("GOT FILE POST")
			if req.FormValue("file") != "" { // upload file
				tmp := strings.Split(req.FormValue("file"), "\\")
				simpleMessage := &SimpleMessage{Contents: tmp[len(tmp)-1], OriginalName: "FILE"}
				ch <- &GossipPacket{Simple: simpleMessage}

			} else {
				// download file
				fmt.Println("GOT FILE DOWNLOAD REQ")
				privateMessage := &PrivateMessage{Destination: req.FormValue("dest"), Text: req.FormValue("metahash"), HopLimit: 10}
				simpleMessage := &SimpleMessage{Contents: req.FormValue("filename"), OriginalName: "FILE_DOWNLOAD"}
				ch <- &GossipPacket{Simple: simpleMessage, Private: privateMessage}
			}

			fmt.Fprintf(w, "succeed")

		default:
			fmt.Println("UNKNOWN REQUEST FROM fileHandler")
		}
	}
}

func p2pDownloadHandler(ch chan *GossipPacket, org string, conn *net.UDPConn, peers *ConcurrentSliceString) func(w http.ResponseWriter, req *http.Request) {

	return func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case "POST":
			fmt.Println("GOT FILE POST")
			if req.FormValue("keywords") != "" { // upload file
				tmp := strings.Split(req.FormValue("keywords"), ",")
				fmt.Println("==== KEYWORDS ====", tmp)

				f1, f2 := newSearchProcess(conn, org, uint64(2), tmp, peers)
				fmt.Println("FILESSSSS:", f1, ",", f2)
				fmt.Fprintf(w, f1+","+f2)
			} else if req.FormValue("download") != "" {
				fileName := req.FormValue("download")
				data := hex.EncodeToString(searchFileTable.Get(fileName).(sFile).MetafileHash[:])

				privateMessage := &PrivateMessage{Destination: "", Text: data, HopLimit: 10}
				simpleMessage := &SimpleMessage{Contents: fileName, OriginalName: "FILE_DOWNLOAD"}
				ch <- &GossipPacket{Simple: simpleMessage, Private: privateMessage}
			}

		default:
			fmt.Println("UNKNOWN REQUEST FROM fileHandler")
		}
	}
}

func nodeHandler(w http.ResponseWriter, req *http.Request) {

	switch req.Method {
	case "POST":
		node := req.FormValue("node")
		if !peer_contained(node) {
			peer_list.Append(node)
			peer_index[node] = peer_list.Len() - 1

			recv_channels[node] = make(chan *StatusPacket)
			go antiEntropyReceiver(gossiper_peer.conn, recv_channels[node], node)
		}

	case "GET":
		// the function of getting peer list
		fmt.Fprintf(w, strings.Join(peer_list.GetSlice(), "<br/>"))

	default:
		fmt.Println("UNKNOWN REQUEST FROM nodeHandler")
	}
}

func idHandler(w http.ResponseWriter, req *http.Request) {

	switch req.Method {
	case "GET":
		// the function of getting peer list
		fmt.Fprintf(w, strings.Join(id_list, "<br/>"))

	default:
		fmt.Println("UNKNOWN REQUEST FROM nodeHandler")
	}
}

func updateIDlist(recv_name string) {
	for _, id := range id_list {
		if id == recv_name {
			return
		}
	}
	if recv_name != *name {
		id_list = append(id_list, recv_name)
	}
	return
}