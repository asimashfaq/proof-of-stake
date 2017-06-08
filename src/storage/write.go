package storage

func WriteBlock(hash [32]byte, encodedBlock []byte) {

	if encodedBlock == nil {
		delete(blocks,hash)
		return
	}
	blocks[hash] = encodedBlock
}

//can't make fixed-size byte, because all tx types go in there
//we'll see later if this is a sensible design choice
func WriteOpenTx(hash [32]byte, encodedTx []byte) {

	if encodedTx == nil {
		delete(opentxs,hash)
		return
	}
	opentxs[hash] = encodedTx
}

func WriteClosedTx(hash [32]byte, encodedTx []byte) {

	closedtxs[hash] = encodedTx
}
