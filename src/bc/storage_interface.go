package bc

import "storage"

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

func readFundsTx(hash [32]byte) (tx *fundsTx) {

	encodedTx := storage.ReadTx(hash)
	if encodedTx == nil {
		return nil
	}
	return DecodeFundsTx(encodedTx)
}

func writeFundsTx(tx *fundsTx) {

	storage.WriteTx(hashFundsTx(tx),EncodeFundsTx(tx))
}

func readAccTx(hash [32]byte) (tx *accTx) {

	encodedTx := storage.ReadTx(hash)
	if encodedTx == nil {
		return nil
	}
	return DecodeAccTx(encodedTx)
}

func writeAccTx(tx *accTx) {

	storage.WriteTx(hashAccTx(tx), EncodeAccTx(tx))
}



func readState(hash [32]byte) (acc *Account) {

	return nil
}

func writeState(acc Account) {


}