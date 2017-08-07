package storage

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/lisgie/bazo_miner/protocol"
	"log"
	"os"
	"time"
)

var (
	db        *bolt.DB
	logger    *log.Logger
	State     = make(map[[32]byte]*protocol.Account)
	RootKeys  = make(map[[32]byte]*protocol.Account)
	txMemPool = make(map[[32]byte]protocol.Transaction)
)

//Entry function for the storage package
func Init(dbname string) {

	LogFile, _ := os.OpenFile("logs/storage "+time.Now().String(), os.O_RDWR|os.O_CREATE, 0666)
	logger = log.New(LogFile, "", log.LstdFlags)

	var err error
	db, err = bolt.Open(dbname, 0600, nil)
	if err != nil {
		logger.Fatal(err)
	}

	db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucket([]byte("openblocks"))
		if err != nil {
			return fmt.Errorf("Create bucket: %s", err)
		}
		return nil
	})
	db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucket([]byte("closedblocks"))
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
