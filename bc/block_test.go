package bc

import (
	"testing"
	"math/rand"
	"time"
	"reflect"
)

//Tests block adding, verification, serialization and deserialization
func TestBlock(t *testing.T) {

	genesis := newBlock()
	lastBlock = genesis
	b := newBlock()
	hashFundsSlice,hashAccSlice := createBlockWithTxs(b)
	b.finalizeBlock()

	encodedBlock := encodeBlock(b)
	decodedBlock := decodeBlock(encodedBlock)

	err := validateBlock(decodedBlock)

	b.stateCopy = nil
	decodedBlock.stateCopy = nil

	if err != nil {
		t.Errorf("Block validation failed (%v)\n", err)
	}
	if !reflect.DeepEqual(hashFundsSlice, decodedBlock.FundsTxData) {
		t.Error("FundsTx data is not properly serialized!")
	}
	if !reflect.DeepEqual(hashAccSlice, decodedBlock.AccTxData) {
		t.Error("AccTx data is not properly serialized!")
	}
	if !reflect.DeepEqual(b, decodedBlock) {
		t.Error("Either serialization or deserialization failed, blocks are not equal!")
	}
}

func TestMultipleBlocks(t *testing.T) {

	genesis := newBlock()
	lastBlock = genesis

	b := newBlock()
	createBlockWithTxs(b)
	b.finalizeBlock()
	if err := validateBlock(b); err != nil {
		t.Errorf("Block failed: %v\n", b)
	}

	b2 := newBlock()
	b2.PrevHash = b.Hash
	createBlockWithTxs(b2)
	b2.finalizeBlock()
	if err := validateBlock(b2); err != nil {
		t.Errorf("Block failed: %v\n", b2)
	}
}

func createBlockWithTxs(b *Block) ([][32]byte, [][32]byte) {

	var testSize uint32
	testSize = 1000

	var hashFundsSlice [][32]byte
	var hashAccSlice [][32]byte
	//in order to create valid funds transactions we need to know the tx count of acc A

	rand := rand.New(rand.NewSource(time.Now().Unix()))
	loopMax := int(rand.Uint32() % testSize)+1
	loopMax += int(accA.TxCnt)
	for cnt := int(accA.TxCnt); cnt < loopMax ; cnt++ {
		tx, _ := ConstrFundsTx(0x01, rand.Uint64()%100+1, rand.Uint64()%100+1, uint32(cnt), accA.Hash, accB.Hash, &PrivKeyA)
		b.addTx(tx)
		hashFundsSlice = append(hashFundsSlice, hashFundsTx(tx))
	}

	loopMax = int(rand.Uint32() % testSize)+1
	for cnt := 0; cnt < loopMax; cnt++ {
		tx, _ := ConstrAccTx(rand.Uint64()%100+1, &RootPrivKey)
		b.addTx(tx)
		hashAccSlice = append(hashAccSlice, hashAccTx(tx))
	}
	return hashFundsSlice,hashAccSlice
}