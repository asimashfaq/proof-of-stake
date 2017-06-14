package storage

import "fmt"

//not accessible from outside
var blocks map[[32]byte][]byte
var opentxs map[[32]byte][]byte
var closedtxs map[[32]byte][]byte

func Init() {
	blocks = make(map[[32]byte][]byte)
	opentxs = make(map[[32]byte][]byte)
	closedtxs = make(map[[32]byte][]byte)
}

//debugging, will be removed later
func DeleteEverything() {
	for key,_ := range blocks {
		delete(blocks,key)
	}
	for key,_ := range opentxs {
		delete(opentxs,key)
	}
	for key,_ := range closedtxs {
		delete(closedtxs,key)
	}
}

func PrintOpenTxs() {
	fmt.Println("OpenTxs:")
	for hash := range opentxs {
		fmt.Printf("%x\n", hash)
	}
	fmt.Println()
}

func PrintClosedTxs() {
	fmt.Println("ClosedTxs:")
	for hash := range closedtxs {
		fmt.Printf("%x\n", hash)
	}
	fmt.Println()
}

func PrintBlocks() {
	fmt.Println("Blocks:")
	for hash := range blocks {
		fmt.Printf("%x\n", hash)
	}
	fmt.Println()
}
