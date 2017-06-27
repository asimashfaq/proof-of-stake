package storage

import (
	"fmt"
	"github.com/boltdb/bolt"
)

func PrintOpenTxs() {

	fmt.Print("Open Txs:\n")
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("opentxs"))
		b.ForEach(func(k, v []byte) error {
			fmt.Printf("%x: %x\n", k, v)
			return nil
		})
		return nil
	})
	fmt.Println()
}

func PrintClosedTxs() {

	fmt.Print("Closed Txs:\n")
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("closedtxs"))
		b.ForEach(func(k, v []byte) error {
			fmt.Printf("%x: %x\n", k, v)
			return nil
		})
		return nil
	})
	fmt.Println()
}

func PrintBlocks() {

	fmt.Print("Blocks:\n")
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("blocks"))
		b.ForEach(func(k, v []byte) error {
			fmt.Printf("%x: %x\n", k, v)
			return nil
		})
		return nil
	})
	fmt.Println()
}
