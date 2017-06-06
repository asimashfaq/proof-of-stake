package bc

import (
	"testing"
	"math/rand"
	"time"
	"reflect"
)

//Tests block adding, verification, serialization and deserialization
func TestBlock(t *testing.T) {

	var testSize uint32
	testSize = 100

	var fundsTxData []*fundsTx
	var accTxData []*accTx
	b := newBlock()

	rand := rand.New(rand.NewSource(time.Now().Unix()))
	loopMax := int(rand.Uint32()%testSize)
	for cnt := 0; loopMax < loopMax; cnt++ {
		tx,_ := ConstrFundsTx(0x01, rand.Uint64()%100+1, rand.Uint64()%100+1, uint32(cnt), accA.Hash, accB.Hash, &PrivKeyA)
		b.addTx(tx)
		fundsTxData = append(fundsTxData, tx)
	}

	loopMax = int(rand.Uint32()%testSize)
	for cnt := 0; cnt < loopMax; cnt++ {
		tx,_ := ConstrAccTx(rand.Uint64()%100+1, &RootPrivKey)
		b.addTx(tx)
		accTxData = append(accTxData, tx)
	}

	b.finalizeBlock()
	encodedBlock := encodeBlock(*b)
	decodedBlock := decodeBlock(encodedBlock)

	err := validateBlock(decodedBlock)

	b.stateCopy = nil
	decodedBlock.stateCopy = nil

	if err != nil {
		t.Errorf("Block validation failed (%v)\n", err)
	}

	if !reflect.DeepEqual(fundsTxData,decodedBlock.FundsTxData) {
		t.Error("FundsTx data is not properly serialized!")
	}

	if !reflect.DeepEqual(accTxData,decodedBlock.AccTxData) {
		t.Error("AccTx data is not properly serialized!")
	}

	if !reflect.DeepEqual(b, decodedBlock) {
		t.Error("Either serialization or deserialization failed, blocks are not equal!")
	}
}