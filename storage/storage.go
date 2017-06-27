package storage

import (
	"github.com/boltdb/bolt"
	"github.com/lisgie/bazo_miner/protocol"
	"fmt"
	"log"
)

var db *bolt.DB
var State map[[8]byte][]*protocol.Account
var RootKeys map[[32]byte]*protocol.Account

func Init() {
	State = make(map[[8]byte][]*protocol.Account)
	RootKeys = make(map[[32]byte]*protocol.Account)

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
		_, err = tx.CreateBucket([]byte("openfunds"))
		if err != nil {
			return fmt.Errorf("Create bucket: %s", err)
		}
		return nil
	})
	db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucket([]byte("openaccs"))
		if err != nil {
			return fmt.Errorf("Create bucket: %s", err)
		}
		return nil
	})
	db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucket([]byte("openconfigs"))
		if err != nil {
			return fmt.Errorf("Create bucket: %s", err)
		}
		return nil
	})
	db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucket([]byte("closedfunds"))
		if err != nil {
			return fmt.Errorf("Create bucket: %s", err)
		}
		return nil
	})
	db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucket([]byte("closedaccs"))
		if err != nil {
			return fmt.Errorf("Create bucket: %s", err)
		}
		return nil
	})
	db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucket([]byte("closedconfigs"))
		if err != nil {
			return fmt.Errorf("Create bucket: %s", err)
		}
		return nil
	})
}

func TearDown() {
	db.Close()
}
