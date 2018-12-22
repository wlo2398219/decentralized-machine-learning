package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"
)

var (
	bcFileMap    = ConcurrentMap{items: map[interface{}]interface{}{}}
	tmpbcFileMap = ConcurrentMap{items: map[interface{}]interface{}{}}
	nextBlock    = Block{PrevHash: [32]byte{}, Transactions: []TxPublish{}}
	// blockChain   = ConcurrentMap{items: map[interface{}]interface{}{}}
	blockChain     = ConcurrentSlice{items: []interface{}{}}
	newBlockChain  = BlockChain{cBlockMap: map[string]*cBlock{}, mainChain: []*cBlock{}}
	parentBlockMap = ConcurrentMap{items: map[interface{}]interface{}{}}
	// chNewBlock = make(chan bool)
	// fileHashMap =
)

type TxPublish struct {
	File     File
	HopLimit uint32
}

type BlockPublish struct {
	Block    Block
	HopLimit uint32
}

type File struct {
	Name         string
	Size         int64
	MetafileHash []byte
}

type Block struct {
	PrevHash     [32]byte
	Nonce        [32]byte
	Transactions []TxPublish
}

type cBlock struct {
	Block  *Block
	Parent *cBlock
	Next   []*cBlock
	Len    int
}

type BlockChain struct {
	sync.RWMutex
	cBlockMap  map[string]*cBlock
	mainChain  []*cBlock
	longestLen int
}

func (b *BlockChain) Get(hash string) *cBlock {
	b.Lock()
	defer b.Unlock()

	return b.cBlockMap[hash]
}

func (b *BlockChain) Set(hash string, c *cBlock) {
	b.Lock()
	defer b.Unlock()

	b.cBlockMap[hash] = c

	return
}

func (b *BlockChain) Exist(hash string) bool {
	b.Lock()
	defer b.Unlock()
	_, exist := b.cBlockMap[hash]
	return exist
}

func (b *BlockChain) Len() int {
	b.Lock()
	defer b.Unlock()

	return len(b.cBlockMap)
}

func (b *Block) Hash() (out [32]byte) {
	h := sha256.New()
	h.Write(b.PrevHash[:])
	h.Write(b.Nonce[:])
	binary.Write(h, binary.LittleEndian, uint32(len(b.Transactions)))
	for _, t := range b.Transactions {
		th := t.Hash()
		h.Write(th[:])
	}
	copy(out[:], h.Sum(nil))
	return
}
func (t *TxPublish) Hash() (out [32]byte) {
	h := sha256.New()
	binary.Write(h, binary.LittleEndian, uint32(len(t.File.Name)))
	h.Write([]byte(t.File.Name))
	h.Write(t.File.MetafileHash)
	copy(out[:], h.Sum(nil))
	return
}

func publishFile(conn *net.UDPConn, f File, peers *ConcurrentSliceString) {
	tx := TxPublish{File: f, HopLimit: 10}
	packet := GossipPacket{TxPublish: &tx}

	tmpbcFileMap.Set(f.Name, f)
	nextBlock.Transactions = append(nextBlock.Transactions, tx)
	// chNewBlock <- true

	// fmt.Println("BLOCK LEN", len(nextBlock.PrevHash), ",", nextBlock.PrevHash)

	for peer := range peers.Iter() {
		sendPacketToAddr(conn, packet, peer)
	}
}

func handlebcFile(conn *net.UDPConn, packet *TxPublish, peers *ConcurrentSliceString, src_addr string) {

	if bcFileMap.Exist(packet.File.Name) || tmpbcFileMap.Exist(packet.File.Name) { // already see this file
		fmt.Println("==== ALREADY SEE THIS FILE ====")
		return
	}

	tmpbcFileMap.Set(packet.File.Name, packet.File)
	nextBlock.Transactions = append(nextBlock.Transactions, *packet)
	// chNewBlock <- true

	if packet.HopLimit == 1 {
		return
	} else {
		packet.HopLimit--
	}

	gPacket := GossipPacket{TxPublish: packet}

	for peer := range peers.Iter() {
		if peer == src_addr {
			continue
		}
		sendPacketToAddr(conn, gPacket, peer)
	}

}

// func printBlockChain() {
// 	slice := blockChain.GetSlice()

// 	fmt.Println("CHAIN")
// 	for i := len(slice) - 1 ; i >= 0 ; i-- {
// 		block := slice[i].(Block)
// 		fmt.Printf("%x:%x:",block.Hash(), block.PrevHash)
// 		if len(block.Transactions) != 0 {
// 			fmt.Printf("%s", block.Transactions[0].File.Name)
// 		}

// 		for j := 1 ; j < len(block.Transactions) ; j++ {
// 			fmt.Printf(",%s", block.Transactions[j].File.Name)
// 		}

// 		fmt.Println()
// 	}
// }

func handleBlock(conn *net.UDPConn, packet *BlockPublish, peers *ConcurrentSliceString, src_addr string) {

	hash := packet.Block.Hash()
	hashHexString := hex.EncodeToString(hash[:])
	phashHexString := hex.EncodeToString(packet.Block.PrevHash[:])
	zeros := [32]byte{}
	zeroString := hex.EncodeToString(zeros[:])

	if hash[0] != 0 || hash[1] != 0 {
		return
	}

	if parentBlockMap.Exist(hash) {
		fmt.Println("===== RECEIVE DUPLICATE BLOCK =====")
		fmt.Printf("PREVHASH: %x\n", packet.Block.PrevHash)
		fmt.Printf("CURRHASH: %x\n", packet.Block.Hash())
		fmt.Println(packet.HopLimit)
		return
	} else {
		fmt.Println("===== RECEIVE NEW BLOCK =====")
		fmt.Printf("PREVHASH: %x\n", packet.Block.PrevHash)
		fmt.Printf("CURRHASH: %x\n", packet.Block.Hash())
		fmt.Println(packet.HopLimit)
	}

	// check the validity
	// 1. no blocks yet, accept any upcoming valid block
	// 2. upcoming block has prevhash = 0
	// 3. prevhash of the upcoming block exists in the blockchain

	if newBlockChain.longestLen == 0 { // accept anyway
		newBlock := &cBlock{Block: &packet.Block, Parent: newBlockChain.Get(zeroString), Next: []*cBlock{}, Len: 1}
		newBlockChain.Set(hashHexString, newBlock)
		newBlockChain.mainChain = append(newBlockChain.mainChain, newBlock)
		newBlockChain.longestLen = 1
		nextBlock.PrevHash = hash
		printNewBlockChain()

	} else if newBlockChain.Exist(phashHexString) {
		parent := newBlockChain.Get(phashHexString)
		curLen := parent.Len + 1
		newBlock := &cBlock{Block: &packet.Block, Parent: parent, Next: []*cBlock{}, Len: curLen}
		newBlockChain.Set(hashHexString, newBlock)
		parent.Next = append(parent.Next, newBlock)
		nextBlock.PrevHash = hash

		fmt.Println("======= CURLEN", curLen, ", LONGEST:", newBlockChain.longestLen, "=======")

		if curLen > newBlockChain.longestLen {
			updateMainChain(newBlock, curLen)
		} else {
			fmt.Println("FORK-SHORTER")
			fmt.Println(hashHexString)
		}
	}
	// if newBlockChain.Len() == 0 || bytes.Equal(packet.Block.PrevHash[:], zeros[:])
	if blockChain.Len() == 0 || bytes.Equal(packet.Block.PrevHash[:], zeros[:]) || parentBlockMap.Exist(packet.Block.PrevHash) {

		// fmt.Println("!!! APPPED NEW BLOCK !!!")
		blockChain.Append(packet.Block)
		// nextBlock.PrevHash = hash
		parentBlockMap.Set(hash, true)
		// printBlockChain()
		// fmt.Println(nextBlock.PrevHash)
	} else {
		// fmt.Println("!!! NEIN !!!")

		return
	}

	if packet.HopLimit == 1 {
		return
	} else {
		packet.HopLimit--
	}

	gPacket := GossipPacket{BlockPublish: packet}

	for peer := range peers.Iter() {
		if peer == src_addr {
			continue
		}
		sendPacketToAddr(conn, gPacket, peer)
	}
}

// func printBlockChain() {
// 	slice := blockChain.GetSlice()

// 	fmt.Println("CHAIN")
// 	for i := len(slice) - 1 ; i >= 0 ; i-- {
// 		block := slice[i].(Block)
// 		fmt.Printf("%x:%x:",block.Hash(), block.PrevHash)
// 		if len(block.Transactions) != 0 {
// 			fmt.Printf("%s", block.Transactions[0].File.Name)
// 		}

// 		for j := 1 ; j < len(block.Transactions) ; j++ {
// 			fmt.Printf(",%s", block.Transactions[j].File.Name)
// 		}

// 		fmt.Println()
// 	}
// }

func printNewBlockChain() {

	fmt.Println("CHAIN")

	for i := newBlockChain.longestLen - 1; i >= 0; i-- {
		block := newBlockChain.mainChain[i]
		fmt.Printf("%x:%x:", block.Block.Hash(), block.Block.PrevHash)

		if len(block.Block.Transactions) != 0 {
			fmt.Printf("%s", block.Block.Transactions[0].File.Name)
		}

		for j := 1; j < len(block.Block.Transactions); j++ {
			fmt.Printf(",%s", block.Block.Transactions[j].File.Name)
		}

		fmt.Println()
	}
}

func updateMainChain(block *cBlock, l int) {
	ind := 0

	tailBlockHash := newBlockChain.mainChain[newBlockChain.longestLen-1].Block.Hash()

	if bytes.Equal(block.Block.PrevHash[:], tailBlockHash[:]) {
		newBlockChain.mainChain = append(newBlockChain.mainChain, block)
		newBlockChain.longestLen = l
		printNewBlockChain()

		return
	}

	// trace back the longest chain now
	chain := make([]*cBlock, l, l)
	tmp := block
	for i := 1; i <= l; i++ {
		chain[l-i] = tmp
		tmp = tmp.Parent
	}

	for ind = 0; ind < newBlockChain.longestLen && newBlockChain.mainChain[ind] == chain[ind]; ind++ {

	}

	numRewind := newBlockChain.longestLen - ind
	fmt.Println("FORK-LONGER rewind", numRewind, "blocks")
	newBlockChain.mainChain = chain
	newBlockChain.longestLen = l
	printNewBlockChain()

	return

}

// Function to calculate the valid block
func PoW(conn *net.UDPConn, peers *ConcurrentSliceString) {
	fmt.Println("????????????????????????????????")
	// ticker := time.NewTicker(time.Millisecond)
	ticker := time.NewTicker(time.Microsecond * 300)
	firstTime := true

	for _ = range ticker.C {

		// randome nonce
		rand.Read(nextBlock.Nonce[:])

		// calculate hash
		out := nextBlock.Hash()

		if out[0] == 0 && out[1] == 0 {

			if firstTime {
				fmt.Println("FIRST BLOCK, WAIT FOR 5 SEC...")
				// time.Sleep(1 * time.Second)
				firstTime = false
			}

			fmt.Printf("FOUND-BLOCK\n%s\n", hex.EncodeToString(out[:]))

			packet := GossipPacket{BlockPublish: &BlockPublish{Block: nextBlock, HopLimit: 20}}
			for peer := range peers.Iter() {
				fmt.Println("SEND BLOCK TO PEER", peer)
				sendPacketToAddr(conn, packet, peer)
			}

			sendPacketToAddr(conn, packet, *gossipAddr)

			// blockChain.Append(nextBlock)
			// parentBlockMap.Set(out, true)
			// nextBlock.PrevHash = out
			// nextBlock.Transactions = []TxPublish{}

			// printNewBlockChain()
			// printBlockChain()
			// time.Sleep(2*time.Second)
		}

	}

}

func initBlockChain() {
	zeros := [32]byte{}
	genesis := &cBlock{Next: []*cBlock{}}
	newBlockChain.Set(hex.EncodeToString(zeros[:]), genesis)
	newBlockChain.longestLen = 0
}
