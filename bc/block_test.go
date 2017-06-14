package bc

import (
	"testing"
	"math/rand"
	"time"
	"reflect"
	"fmt"
)

//Tests block adding, verification, serialization and deserialization
func TestBlock(t *testing.T) {

	cleanAndPrepare()

	b := newBlock()
	hashFundsSlice,hashAccSlice,hashConfigSlice := createBlockWithTxs(b)
	b.finalizeBlock()

	encodedBlock := encodeBlock(b)
	fmt.Printf("%v\n", b)
	decodedBlock := decodeBlock(encodedBlock)
	fmt.Printf("%v\n", decodedBlock)
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
	if !reflect.DeepEqual(hashConfigSlice, decodedBlock.ConfigTxData) {
		fmt.Printf("%v, %v\n", len(hashConfigSlice), len(decodedBlock.ConfigTxData))
		t.Error("ConfigTx data is not properly serialized!")
	}
	if !reflect.DeepEqual(b, decodedBlock) {
		t.Error("Either serialization or deserialization failed, blocks are not equal!")
	}
}

func TestMultipleBlocks(t *testing.T) {

	cleanAndPrepare()
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

func createBlockWithTxs(b *Block) ([][32]byte, [][32]byte, [][32]byte) {

	var testSize uint32
	testSize = 1000

	var hashFundsSlice [][32]byte
	var hashAccSlice [][32]byte
	var hashConfigSlice [][32]byte
	//in order to create valid funds transactions we need to know the tx count of acc A

	rand := rand.New(rand.NewSource(time.Now().Unix()))
	loopMax := int(rand.Uint32() % testSize)+1
	loopMax += int(accA.TxCnt)
	for cnt := int(accA.TxCnt); cnt < loopMax ; cnt++ {
		accAHash := serializeHashContent(accA.Address)
		accBHash := serializeHashContent(accB.Address)
		tx, _ := ConstrFundsTx(0x01, rand.Uint64()%100+1, rand.Uint64()%100+1, uint32(cnt), accAHash, accBHash, &PrivKeyA)
		if err := b.addTx(tx); err == nil {
			hashFundsSlice = append(hashFundsSlice, hashFundsTx(tx))
		}
	}

	loopMax = int(rand.Uint32() % testSize)+1
	for cnt := 0; cnt < loopMax; cnt++ {
		tx, _ := ConstrAccTx(rand.Uint64()%100+1, &RootPrivKey)
		if err := b.addTx(tx); err == nil {
			hashAccSlice = append(hashAccSlice, hashAccTx(tx))
		}
	}

	loopMax = int(rand.Uint32() % testSize)+1
	for cnt := 0; cnt < loopMax; cnt++ {
		tx,_:= ConstrConfigTx(uint8(rand.Uint32()%256), uint8(rand.Uint32()%5+1),rand.Uint64(), rand.Uint64(), &RootPrivKey)
		if err := b.addTx(tx); err == nil {
			hashConfigSlice = append(hashConfigSlice, hashConfigTx(tx))
		}
	}

	return hashFundsSlice,hashAccSlice,hashConfigSlice
}