package storage

import (
	"github.com/boltdb/bolt"
	"github.com/lisgie/bazo_miner/protocol"
)

//TODO: Error checking
func WriteOpenBlock(block *protocol.Block) {

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("openblocks"))
		err := b.Put(block.Hash[:], block.Encode())
		return err
	})
}

func WriteClosedBlock(block *protocol.Block) {

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("closedblocks"))
		err := b.Put(block.Hash[:], block.Encode())
		return err
	})
}

//breaking the "tx" shortcut for here and using "transaction" to distinguish between bolt's transactions
func WriteOpenTx(transaction protocol.Transaction) {

	txMemPool[transaction.Hash()] = transaction

	/*var bucket string
	switch transaction.(type) {
	case *protocol.FundsTx:
		bucket = "openfunds"
	case *protocol.AccTx:
		bucket = "openaccs"
	case *protocol.ConfigTx:
		bucket = "openconfigs"
	}

	hash := transaction.Hash()
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		err := b.Put(hash[:], transaction.Encode())
		return err
	})*/
}

func WriteClosedTx(transaction protocol.Transaction) {

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
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		err := b.Put(hash[:], transaction.Encode())
		return err
	})
}
