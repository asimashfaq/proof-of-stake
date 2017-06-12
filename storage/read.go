package storage

func ReadBlock(hash [32]byte) (encodedBlock []byte) {

	if block,exists := blocks[hash]; exists {
		return block[:]
	}
	return nil
}

func ReadOpenTx(hash [32]byte) (encodedTx []byte) {

	if tx,exists := opentxs[hash]; exists {
		return tx[:]
	}
	return nil
}

func ReadClosedTx(hash [32]byte) (encodedTx []byte) {

	if tx,exists := closedtxs[hash]; exists {
		return tx[:]
	}
	return nil
}