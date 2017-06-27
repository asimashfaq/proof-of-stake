package storage

import (
	"github.com/lisgie/bazo_miner/protocol"
	"github.com/boltdb/bolt"
)

func ReadBlock(hash [32]byte) (block *protocol.Block) {

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

func ReadOpenTx(hash [32]byte) (transaction protocol.Transaction) {

	var encodedTx []byte
	var fundstx *protocol.FundsTx
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("openfunds"))
		encodedTx = b.Get(hash[:])
		return nil
	})
	if encodedTx != nil {
		return fundstx.Decode(encodedTx)
	}

	var acctx *protocol.AccTx
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("openaccs"))
		encodedTx = b.Get(hash[:])
		return nil
	})
	if encodedTx != nil {
		return acctx.Decode(encodedTx)
	}

	var configtx *protocol.ConfigTx
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("openconfigs"))
		encodedTx = b.Get(hash[:])
		return nil
	})
	if encodedTx != nil {
		return configtx.Decode(encodedTx)
	}
	return nil
}

//needed for the miner to prepare a new block
func ReadAllOpenTxs() (allOpenTxs []protocol.Transaction) {

	var fundstx *protocol.FundsTx
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("openfunds"))
		b.ForEach(func(k, v []byte) error {
			allOpenTxs = append(allOpenTxs,fundstx.Decode(v))
			return nil
		})
		return nil
	})

	var acctx *protocol.AccTx
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("openaccs"))
		b.ForEach(func(k, v []byte) error {
			allOpenTxs = append(allOpenTxs,acctx.Decode(v))
			return nil
		})
		return nil
	})

	var configtx *protocol.ConfigTx
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("openconfigs"))
		b.ForEach(func(k, v []byte) error {
			allOpenTxs = append(allOpenTxs,configtx.Decode(v))
			return nil
		})
		return nil
	})

	return allOpenTxs
}

func ReadClosedTx(hash [32]byte) (transaction protocol.Transaction) {

	var encodedTx []byte
	var fundstx *protocol.FundsTx
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("closedfunds"))
		encodedTx = b.Get(hash[:])
		return nil
	})
	if encodedTx != nil {
		return fundstx.Decode(encodedTx)
	}

	var acctx *protocol.AccTx
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("closedaccs"))
		encodedTx = b.Get(hash[:])
		return nil
	})
	if encodedTx != nil {
		return acctx.Decode(encodedTx)
	}

	var configtx *protocol.ConfigTx
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("closedconfigs"))
		encodedTx = b.Get(hash[:])
		return nil
	})
	if encodedTx != nil {
		return configtx.Decode(encodedTx)
	}
	return nil
}
