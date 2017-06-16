package storage

import (
	"fmt"
	"github.com/boltdb/bolt"
	"log"
)

//not accessible from outside
var blocks map[[32]byte][]byte
var opentxs map[[32]byte][]byte
var closedtxs map[[32]byte][]byte

var db *bolt.DB

func Init() {

	var err error
	db, err = bolt.Open("bc.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucket([]byte("bc"))
		if err != nil {
			return fmt.Errorf("Create bucket: %s", err)
		}
		return nil
	})

	blocks = make(map[[32]byte][]byte)
	opentxs = make(map[[32]byte][]byte)
	closedtxs = make(map[[32]byte][]byte)
}

//debugging, will be removed later
func DeleteEverything() {
	for key, _ := range blocks {
		delete(blocks, key)
	}
	for key, _ := range opentxs {
		delete(opentxs, key)
	}
	for key, _ := range closedtxs {
		delete(closedtxs, key)
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
