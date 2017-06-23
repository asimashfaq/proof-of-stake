package storage

import (
	"github.com/boltdb/bolt"
)

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
