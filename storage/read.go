package storage

import (
	"github.com/boltdb/bolt"
)

func ReadBlock(hash [32]byte) (encodedBlock []byte) {

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("bc"))
		encodedBlock = b.Get(hash[:])
		return nil
	})

	if encodedBlock == nil {
		return nil
	}

	return encodedBlock

	if block, exists := blocks[hash]; exists {
		return block[:]
	}
	return nil
}

func ReadOpenTx(hash [32]byte) (encodedTx []byte) {

	if tx, exists := opentxs[hash]; exists {
		return tx[:]
	}
	return nil
}

func ReadClosedTx(hash [32]byte) (encodedTx []byte) {

	if tx, exists := closedtxs[hash]; exists {
		return tx[:]
	}
	return nil
}
