package miner

import (
	"github.com/lisgie/bazo_miner/protocol"
	"github.com/boltdb/bolt"
	"fmt"
	"log"
)

var db *bolt.DB
var State map[[8]byte][]*protocol.Account
var RootKeys map[[32]byte]*protocol.Account

func storageInit() {
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

func storageTearDown() {
	db.Close()
}

//READS------------------------------------------------


//acts as the interface to the storage module

func readBlock(hash [32]byte) (block *protocol.Block) {

	var encodedBlock []byte
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("blocks"))
		encodedBlock = b.Get(hash[:])
		return nil
	})

	if encodedBlock == nil {
		return nil
	}

	return block.Decode(encodedBlock)
}

func ReadOpenTx(hash [32]byte) (encodedTx []byte) {

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("opentxs"))
		encodedTx = b.Get(hash[:])
		return nil
	})

	if encodedTx == nil {
		return nil
	}

	return encodedTx
}

func ReadClosedTx(hash [32]byte) (encodedTx []byte) {

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("closedtxs"))
		encodedTx = b.Get(hash[:])
		return nil
	})

	if encodedTx == nil {
		return nil
	}

	return encodedTx
}

//WRITES-----------------------------------------------

//if the data has already been written, nil is returned, this is important to check
func WriteBlock(hash [32]byte, encodedBlock []byte) {

	if encodedBlock == nil {
		db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("blocks"))
			err := b.Delete(hash[:])
			return err
		})
	} else {
		db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("blocks"))
			err := b.Put(hash[:], encodedBlock)
			return err
		})
	}
}

//can't make fixed-size byte, because all tx types go in there
//we'll see later if this is a sensible design choice
func WriteOpenTx(hash [32]byte, encodedTx []byte) {

	if encodedTx == nil {
		db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("opentxs"))
			err := b.Delete(hash[:])
			return err
		})
	} else {
		db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("opentxs"))
			err := b.Put(hash[:], encodedTx)
			return err
		})
	}
}

func WriteClosedTx(hash [32]byte, encodedTx []byte) {

	if encodedTx == nil {
		db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("closedtxs"))
			err := b.Delete(hash[:])
			return err
		})
	} else {
		db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("closedtxs"))
			err := b.Put(hash[:], encodedTx)
			return err
		})
	}
}

//DELETES----------------------------------------------

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

//DEBUGGING--------------------------------------------


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





























func writeBlock(b *protocol.Block) {

	storage.WriteBlock(b.Hash, b.Encode())
}

func readOpenFundsTx(hash [32]byte) (tx *protocol.FundsTx) {

	encodedTx := storage.ReadOpenTx(hash)
	if encodedTx == nil {
		return nil
	}
	decodedTx := tx.Decode(encodedTx)
	//this isn't very nice, but necessary to enrich the full hashes in the transaction
	verifyFundsTx(decodedTx)
	return decodedTx

}

func readClosedFundsTx(hash [32]byte) (tx *protocol.FundsTx) {

	encodedTx := storage.ReadClosedTx(hash)
	if encodedTx == nil {
		return nil
	}
	decodedTx := tx.Decode(encodedTx)
	verifyFundsTx(decodedTx)
	return decodedTx
}

func readOpenAccTx(hash [32]byte) (tx *protocol.AccTx) {

	encodedTx := storage.ReadOpenTx(hash)
	if encodedTx == nil {
		return nil
	}
	return tx.Decode(encodedTx)
}

func readClosedAccTx(hash [32]byte) (tx *protocol.AccTx) {

	encodedTx := storage.ReadClosedTx(hash)
	if encodedTx == nil {
		return nil
	}
	return tx.Decode(encodedTx)
}

func readOpenConfigTx(hash [32]byte) (tx *protocol.ConfigTx) {

	encodedTx := storage.ReadOpenTx(hash)
	if encodedTx == nil {
		return nil
	}
	return tx.Decode(encodedTx)
}

func readClosedConfigTx(hash [32]byte) (tx *protocol.ConfigTx) {

	encodedTx := storage.ReadClosedTx(hash)
	if encodedTx == nil {
		return nil
	}
	return tx.Decode(encodedTx)
}

func writeOpenTx(tx protocol.Transaction) {

	storage.WriteOpenTx(tx.Hash(), tx.Encode())
}

func writeClosedTx(tx protocol.Transaction) {

	storage.WriteClosedTx(tx.Hash(), tx.Encode())

}

func readAllOpenTxs() []storage.Kvtuple { return storage.ReadAllOpenTxs() }

func deleteOpenTx(hash [32]byte) {

	storage.WriteOpenTx(hash, nil)
}

//delete in the closed confirmation is needed as well, in case of block rollback
func deleteClosedTx(hash [32]byte) {

	storage.WriteClosedTx(hash, nil)
}

func deleteBlock(hash [32]byte) {
	storage.WriteBlock(hash, nil)
}

func deleteAll() {
	storage.DeleteAll()
}
