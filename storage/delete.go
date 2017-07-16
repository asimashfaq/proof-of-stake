package storage

import (
	"github.com/boltdb/bolt"
	"github.com/lisgie/bazo_miner/protocol"
)

func DeleteOpenBlock(hash [32]byte) {

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("openblocks"))
		err := b.Delete(hash[:])
		return err
	})
}

func DeleteClosedBlock(hash [32]byte) {

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("closedblocks"))
		err := b.Delete(hash[:])
		return err
	})
}

func DeleteOpenTx(transaction protocol.Transaction) {

	delete(txMemPool, transaction.Hash())

	/*var bucket string
	switch transaction.(type) {
	case *protocol.FundsTx:
		bucket = "openfunds"
	case *protocol.AccTx:
		bucket = "openaccs"
	case *protocol.ConfigTx:
		bucket = "openconfigs"
	}

	//get (slice of unaddressable value) error if we do it directly
	hash := transaction.Hash()
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		err := b.Delete(hash[:])
		return err
	})*/
}

func DeleteClosedTx(transaction protocol.Transaction) {

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
		err := b.Delete(hash[:])
		return err
	})
}

func DeleteAll() {

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("openblocks"))
		b.ForEach(func(k, v []byte) error {
			b.Delete(k)
			return nil
		})
		return nil
	})
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("closedblocks"))
		b.ForEach(func(k, v []byte) error {
			b.Delete(k)
			return nil
		})
		return nil
	})
	/*db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("openfunds"))
		b.ForEach(func(k, v []byte) error {
			b.Delete(k)
			return nil
		})
		return nil
	})
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("openaccs"))
		b.ForEach(func(k, v []byte) error {
			b.Delete(k)
			return nil
		})
		return nil
	})
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("openconfigs"))
		b.ForEach(func(k, v []byte) error {
			b.Delete(k)
			return nil
		})
		return nil
	})*/
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("closedfunds"))
		b.ForEach(func(k, v []byte) error {
			b.Delete(k)
			return nil
		})
		return nil
	})
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("closedaccs"))
		b.ForEach(func(k, v []byte) error {
			b.Delete(k)
			return nil
		})
		return nil
	})
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("closedconfigs"))
		b.ForEach(func(k, v []byte) error {
			b.Delete(k)
			return nil
		})
		return nil
	})
}
