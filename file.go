package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

type mFile struct {
	FileName       string
	ChunkCount     uint64
	AvailableIndex []uint64
	MetaFile       []byte
	MetaHash       []byte
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

type downloadChannel struct {
	MetaHashString string
	MetaFileString string
	Ch             chan *DataReply
}

var (
	// hashDataTable = map[string]FileChunck{}
	hashSearchMap = map[string]bool{}
	downloadTable = ConcurrentMap{items: map[interface{}]interface{}{}}
)

func handleDataReplies(packet *DataReply) {
	// hexString := hex.EncodeToString(packet.HashValue)
	hashString := string(packet.HashValue)

	for m := range downloadTable.Iter() {

		if m.(downloadChannel).MetaHashString == hashString {
			downloadTable.Set(m.(downloadChannel).MetaHashString, downloadChannel{MetaHashString: m.(downloadChannel).MetaHashString, MetaFileString: string(packet.Data), Ch: m.(downloadChannel).Ch})
			m.(downloadChannel).Ch <- packet
		} else if strings.Contains(m.(downloadChannel).MetaFileString, hashString) {
			m.(downloadChannel).Ch <- packet
		}
	}
}

// func GetFile(fileName string, shareFiles map[string]bool, fileChunkIndex map[string]uint64) mFile {
func GetFile(fileName string, fileData map[string]*mFile) File {

	const CHUNK_SIZE = 8192
	var (
		N      uint64
		indArr = []uint64{}
	)

	content, err := ioutil.ReadFile("./_SharedFiles/" + fileName)
	if err != nil {
		log.Fatal(err)
	}

	if len(content)%CHUNK_SIZE == 0 {
		N = uint64(len(content)) / CHUNK_SIZE
	} else {
		N = uint64(len(content)/CHUNK_SIZE) + 1
	}

	metafile := make([]byte, N*32)

	for i := uint64(0); i < N-1; i++ {
		sum := sha256.Sum256(content[i*CHUNK_SIZE : (i+1)*CHUNK_SIZE])
		storeChunk(sum[:], content[i*CHUNK_SIZE:(i+1)*CHUNK_SIZE])

		for j := uint64(0); j < 32; j++ {
			metafile[i*32+j] = sum[j]
		}
	}

	sum := sha256.Sum256(content[(N-1)*CHUNK_SIZE:])
	storeChunk(sum[:], content[(N-1)*CHUNK_SIZE:])

	for j := uint64(0); j < 32; j++ {
		metafile[(N-1)*32+j] = sum[j]
	}

	shameta := sha256.Sum256(metafile)
	storeChunk(shameta[:], metafile)

	for i := uint64(1); i <= N; i++ {
		indArr = append(indArr, i)
	}

	// fileData[fileName] = mFile{FileName: fileName, ChunkCount: uint64(N), MetaFile: metafile, MetaHash: shameta[:], AvailableIndex: uint64(N)}
	fileData[fileName] = &mFile{FileName: fileName, ChunkCount: uint64(N), MetaFile: metafile, MetaHash: shameta[:], AvailableIndex: indArr}

	return File{Name: fileName, MetafileHash: shameta[:], Size: int64(len(content))}
}

func PrintFileInfo(f mFile) {
	fmt.Println("File NAME: " + f.FileName)
	fmt.Println("File META:", f.MetaFile)
	fmt.Printf("META HASH:%x", f.MetaHash)
}

func storeChunk(hashValue, data []byte) {

	hashString := hex.EncodeToString(hashValue)
	hashSearchMap[hashString] = true

	fptr, _ := os.Create("./_Chunks/" + hashString)
	_, _ = fptr.Write(data)
	fptr.Close()
}

func DownloadFile(conn *net.UDPConn, fileName, src string, dst interface{}, metaHash []byte, fileData map[string]*mFile) {
	var (
		fileHash []byte
		// mFile     []byte
		dstArr = []string{}
		fptr   *os.File
	)

	gotMeta, notFound := false, false
	index, total := uint64(0), uint64(0)
	strMetaHash := string(metaHash)

	fmt.Println("====== DOWNLOAD FILE =====", fileName)

	switch dst.(type) {
	case string:
		dstArr = append(dstArr, dst.(string))
	case []string:
		for _, d := range dst.([]string) {
			dstArr = append(dstArr, d)
		}
	}

	// create a new downloading record
	data := make(chan *DataReply)
	downloadTable.Set(strMetaHash, downloadChannel{MetaHashString: strMetaHash, Ch: data})

	// send metadata request
	metaRequest := DataRequest{Origin: src, Destination: dstArr[0], HopLimit: 10, HashValue: metaHash}
	_ = sendPacketToAddr(conn, GossipPacket{DataRequest: &metaRequest}, nextHopTable[dstArr[0]].NextHop)

	for {
		timer := time.NewTimer(time.Second * 5)
		select {
		case d := <-data:
			if len(d.HashValue) == 0 { // handle the destination doesn't have the file
				notFound = true
				fmt.Println("~~~~~~~~~ DESTINATION DOESN'T HAVE THE File", fileName, "~~~~~~~~~")
				break
			} else if !gotMeta && strMetaHash == string(d.HashValue) { // get metafilehash
				gotMeta = true
				fileHash = make([]byte, len(d.Data))
				copy(fileHash, d.Data)

				total = uint64(len(fileHash)) / 32

				if total >= 1 && len(dstArr) == 1 {
					for i := uint64(1); i < total; i++ {
						dstArr = append(dstArr, dstArr[0])
					}
				}

				fileData[fileName] = &mFile{FileName: fileName, ChunkCount: uint64(total), MetaFile: fileHash, MetaHash: metaHash, AvailableIndex: []uint64{}}

				storeChunk(metaHash, fileHash)

				// ask for the 1st chunk
				metaRequest := DataRequest{Origin: src, Destination: dstArr[0], HopLimit: 10, HashValue: fileHash[:32]}
				_ = sendPacketToAddr(conn, GossipPacket{DataRequest: &metaRequest}, nextHopTable[dstArr[0]].NextHop)
				fptr, _ = os.Create("./_Downloads/" + fileName)

				break

			} else if gotMeta && string(fileHash[index*32:(index+1)*32]) == string(d.HashValue) {
				index++
				fmt.Println("DOWNLOADING", fileName, "chunck", index, "from", dstArr[index-1])

				_, _ = fptr.Write(d.Data)

				storeChunk(d.HashValue, d.Data)
				fileData[fileName].AvailableIndex = append(fileData[fileName].AvailableIndex, index)

				if total == index {
					// close the channel and return
					fmt.Println("RECONSTRUCTED File", fileName)
					fptr.Close()
					return
				} else {
					metaRequest := DataRequest{Origin: src, Destination: dstArr[index], HopLimit: 10, HashValue: fileHash[index*32 : (index+1)*32]}
					_ = sendPacketToAddr(conn, GossipPacket{DataRequest: &metaRequest}, nextHopTable[dstArr[index]].NextHop)
					break
				}
			} else {
				fmt.Println(d.HashValue)
				fmt.Println("I DON'T UNDERSTAND!!!!!!!")
			}

		case <-timer.C:
			// retransmit
			fmt.Println("File DOWNLOAD TIMEOUT")
			if !gotMeta {
				metaRequest := DataRequest{Origin: src, Destination: dstArr[0], HopLimit: 10, HashValue: metaHash}
				_ = sendPacketToAddr(conn, GossipPacket{DataRequest: &metaRequest}, nextHopTable[dstArr[0]].NextHop)
			} else {
				metaRequest := DataRequest{Origin: src, Destination: dstArr[index], HopLimit: 10, HashValue: fileHash[index*32 : (index+1)*32]}
				_ = sendPacketToAddr(conn, GossipPacket{DataRequest: &metaRequest}, nextHopTable[dstArr[index]].NextHop)
			}

		}

		if notFound {
			break
		}
	}

	return
}

func searchHashAndSend(conn *net.UDPConn, hashValue []byte, src, dst string) {
	hashString := hex.EncodeToString(hashValue)

	if _, exist := hashSearchMap[hashString]; exist {
		data, err := ioutil.ReadFile("./_Chunks/" + hashString)

		if err != nil {
			fmt.Println("File READING ERROR IN searchHashAndSend")
		}

		metaReply := DataReply{Origin: src, Destination: dst, HopLimit: 10, HashValue: hashValue, Data: data}
		_ = sendPacketToAddr(conn, GossipPacket{DataReply: &metaReply}, nextHopTable[dst].NextHop)
		return
	} else {
		fmt.Println("DOESNT' EXIST")
	}

	// send empty hash to notify we don't have this chunck
	fmt.Println("------------------ DON'T HAVE THIS File, SEND NIL REPLY ----------------------------")
	metaReply := DataReply{Origin: src, Destination: dst, HopLimit: 10, HashValue: nil}
	_ = sendPacketToAddr(conn, GossipPacket{DataReply: &metaReply}, nextHopTable[dst].NextHop)

	return
}
