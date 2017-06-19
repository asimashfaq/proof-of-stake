package miner

import (
	"github.com/lisgie/bazo_miner/protocol"
	"github.com/lisgie/bazo_miner/storage"
)

//acts as the interface to the storage module
func readBlock(hash [32]byte) (b *protocol.Block) {

	encodedBlock := storage.ReadBlock(hash)
	if encodedBlock == nil {
		return nil
	}

	return b.Decode(encodedBlock)
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
