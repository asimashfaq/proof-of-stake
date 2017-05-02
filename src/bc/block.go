package bc

import "hash"

type Block struct {
	hash hash.Hash32
	prevHash hash.Hash32
	timestamp int64
	nrOfTransactions int
	data []Transaction
}

func (b *Block) addTransact(t Transaction) {

}

func (b *Block) finalizeBlock(){

}