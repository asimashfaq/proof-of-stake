package storage

import (
	"fmt"
	"github.com/boltdb/bolt"
	"log"
)

var db *bolt.DB

func Init() {

	var err error
	db, err = bolt.Open("miner.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucket([]byte("blocks"))
		if err != nil {
			return fmt.Errorf("Create bucket: %s", err)
		}
		return nil
	})
	db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucket([]byte("opentxs"))
		if err != nil {
			return fmt.Errorf("Create bucket: %s", err)
		}
		return nil
	})
	db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucket([]byte("closedtxs"))
		if err != nil {
			return fmt.Errorf("Create bucket: %s", err)
		}
		return nil
	})
}

func TearDown() {
	db.Close()
}

//debugging, will be removed later
func DeleteAll() {

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("blocks"))
		b.ForEach(func(k, v []byte) error {
			b.Delete(k)
			return nil
		})
		return nil
	})

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("opentxs"))
		b.ForEach(func(k, v []byte) error {
			b.Delete(k)
			return nil
		})
		return nil
	})
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("closedtxs"))
		b.ForEach(func(k, v []byte) error {
			b.Delete(k)
			return nil
		})
		return nil
	})
}

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
