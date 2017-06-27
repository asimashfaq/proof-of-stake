package storage

import (
	"github.com/boltdb/bolt"
	"github.com/lisgie/bazo_miner/protocol"
)

func WriteBlock(block *protocol.Block) {

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("blocks"))
		err := b.Put(block.Hash[:], block.Encode())
		return err
	})
}

//breaking the "tx" shortcut for here and using "transaction" to distinguish between bolt's transactions
func WriteOpenTx(transaction protocol.Transaction) {

	var bucket string
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
	})
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
