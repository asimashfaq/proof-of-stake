package storage

func WriteBlock(hash [32]byte, block []byte) {

}

//can't make fixed-size byte, because all tx types go in there
//we'll see later if this is a sensible design choice
func WriteTx(hash [32]byte, encodedTx []byte) {

	txs[hash] = encodedTx
}
