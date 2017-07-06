package storage

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/lisgie/bazo_miner/protocol"
)

func PrintOpenTxs() {

	fmt.Print("Open Txs:\n")
	fmt.Print("Open Funds:\n")
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("openfunds"))
		b.ForEach(func(k, v []byte) error {
			fmt.Printf("%x: %x\n", k, v)
			return nil
		})
		return nil
	})
	fmt.Print("Open Accs:\n")
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("openaccs"))
		b.ForEach(func(k, v []byte) error {
			fmt.Printf("%x: %x\n", k, v)
			return nil
		})
		return nil
	})
	fmt.Print("Open Configs:\n")
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("openconfigs"))
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
	fmt.Print("Closed Funds:\n")
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("closedfunds"))
		b.ForEach(func(k, v []byte) error {
			var ftx *protocol.FundsTx
			fmt.Printf("%x: %v\n", k, ftx.Decode(v))
			return nil
		})
		return nil
	})
	fmt.Print("Closed Accs:\n")
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("closedaccs"))
		b.ForEach(func(k, v []byte) error {
			fmt.Printf("%x: %x\n", k, v)
			return nil
		})
		return nil
	})
	fmt.Print("Closed Configs:\n")
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("closedconfigs"))
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
