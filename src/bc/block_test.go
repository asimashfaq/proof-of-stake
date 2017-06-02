package bc

import (
	"testing"
	"math/rand"
	"time"
	"reflect"
)

func TestSerialization(t *testing.T) {

	var fundsTxData []fundsTx
	var accTxData []accTx
	b := newBlock()

	rand := rand.New(rand.NewSource(time.Now().Unix()))
	loopMax := int(rand.Uint32()%1000)
	for cnt := 0; cnt < loopMax; cnt++ {
		tx,_ := ConstrFundsTx(0x01, rand.Uint64()%100+1, rand.Uint64()%100+1, uint32(cnt), accA.Hash, accB.Hash, &PrivKeyA)
		b.addTx(&tx)
		fundsTxData = append(fundsTxData, tx)
	}

	loopMax = int(rand.Uint32()%1000)
	for cnt := 0; cnt < loopMax; cnt++ {
		tx,_ := ConstrAccTx(rand.Uint64()%123435, &RootPrivKey)
		b.addTx(&tx)
		accTxData = append(accTxData, tx)
	}

	b.finalizeBlock()
	encodedBlock := encodeBlock(*b)
	decodedBlock := decodeBlock(encodedBlock)

	validateBlock(decodedBlock)

	b.stateCopy = nil
	decodedBlock.stateCopy = nil

	if reflect.DeepEqual(b, decodedBlock) == false {
		t.Error("Either serialization or deserialization failed, blocks are not equal!")
	}
}