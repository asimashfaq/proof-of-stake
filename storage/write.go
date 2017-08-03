package storage

import (
	"github.com/boltdb/bolt"
	"github.com/lisgie/bazo_miner/protocol"
)

func WriteOpenBlock(block *protocol.Block) (err error) {

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("openblocks"))
		err := b.Put(block.Hash[:], block.Encode())
		return err
	})

	return err
}

func WriteClosedBlock(block *protocol.Block) (err error) {

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("closedblocks"))
		err := b.Put(block.Hash[:], block.Encode())
		return err
	})

	return err
}

//Changing the "tx" shortcut here and using "transaction" to distinguish between bolt's transactions
func WriteOpenTx(transaction protocol.Transaction) {

	txMemPool[transaction.Hash()] = transaction
}

func WriteClosedTx(transaction protocol.Transaction) (err error) {

	var bucket string
	switch transaction.(type) {
	case *protocol.FundsTx:
		bucket = "closedfunds"
	case *protocol.AccTx:
		bucket = "closedaccs"
	case *protocol.ConfigTx:
		bucket = "closedconfigs"
	}

	hash := transaction.Hash()
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		err := b.Put(hash[:], transaction.Encode())
		return err
	})

	return err
}
