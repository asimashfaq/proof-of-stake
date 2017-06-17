package storage

import (
	"github.com/boltdb/bolt"
)

func ReadBlock(hash [32]byte) (encodedBlock []byte) {

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("blocks"))
		encodedBlock = b.Get(hash[:])
		return nil
	})

	if encodedBlock == nil {
		return nil
	}

	return encodedBlock
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
