package bc

import (
	"testing"
	"fmt"
	"math/rand"
)

func TestAddingTxs(t *testing.T) {

}

func TestSerialization(t *testing.T) {

	b := newBlock()
	tx1,_ := ConstrFundsTx(0x01, rand.Uint64()%100+1, rand.Uint64()%100+1, 0, accA.Hash, accB.Hash, &PrivKeyA)

	b.addTx(&tx1)
	b.finalizeBlock()
	encodedBlock := encodeBlock(*b)

	fmt.Printf("%v\n", b)

	fmt.Printf("%x\n", encodedBlock)
	fmt.Printf("%v\n", decodeBlock(encodedBlock))
}