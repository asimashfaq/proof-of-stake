package bc

import (
	"storage"
	"bytes"
	"fmt"
)

//acts as the interface to the storage module
func readBlock(hash [32]byte) (b *Block) {

	encodedBlock := storage.ReadBlock(hash)
	if encodedBlock == nil {
		return nil
	}
	return decodeBlock(encodedBlock)
}

func writeBlock(b *Block) {

	storage.WriteBlock(b.Hash,encodeBlock(b))
}

func readOpenFundsTx(hash [32]byte) (tx *fundsTx) {

	encodedTx := storage.ReadOpenTx(hash)
	if encodedTx == nil {
		return nil
	}
	decodedTx := DecodeFundsTx(encodedTx)
	//this isn't very nice, but necessary to enrich the full hashes in the transaction
	decodedTx.verify()
	return decodedTx

}

func readClosedFundsTx(hash [32]byte) (tx *fundsTx) {

	encodedTx := storage.ReadClosedTx(hash)
	if encodedTx == nil {
		return nil
	}
	decodedTx := DecodeFundsTx(encodedTx)
	decodedTx.verify()
	return decodedTx
}

func readOpenAccTx(hash [32]byte) (tx *accTx) {

	encodedTx := storage.ReadOpenTx(hash)
	if encodedTx == nil {
		return nil
	}
	return DecodeAccTx(encodedTx)
}

func readClosedAccTx(hash [32]byte) (tx *accTx) {

	encodedTx := storage.ReadClosedTx(hash)
	if encodedTx == nil {
		return nil
	}
	return DecodeAccTx(encodedTx)
}
func getAccountFromHash(hash [32]byte) (*Account) {

	var fixedHash [8]byte
	copy(fixedHash[:],hash[0:8])
	for _,acc := range State[fixedHash] {
		if bytes.Compare(acc.Hash[:],hash[:]) == 0 {
			return acc
		}
	}
	return nil
}
func readState(hash [32]byte) (acc *Account) {

	var shortHash [8]byte
	copy(shortHash[:],hash[0:8])
	accSlice := storage.ReadState(shortHash)

	//decode the Slice
	for _,decodedAcc := range accSlice {
		fmt.Printf("%x\n", decodedAcc)
	}
	return nil
}

func writeOpenFundsTx(tx *fundsTx) {

	storage.WriteOpenTx(hashFundsTx(tx),EncodeFundsTx(tx))
}

func writeClosedFundsTx(tx *fundsTx) {

	storage.WriteClosedTx(hashFundsTx(tx),EncodeFundsTx(tx))
}

func writeOpenAccTx(tx *accTx) {

	storage.WriteOpenTx(hashAccTx(tx),EncodeAccTx(tx))
}

func writeClosedAccTx(tx *accTx) {

	storage.WriteClosedTx(hashAccTx(tx),EncodeAccTx(tx))
}

func writeState(acc *Account) {}

func deleteOpenFundsTx(hash [32]byte) {

	storage.WriteOpenTx(hash,nil)
}

func deleteOpenAccTx(hash [32]byte) {

	storage.WriteOpenTx(hash,nil)
}

