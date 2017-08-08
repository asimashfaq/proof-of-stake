package miner

import (
	"github.com/lisgie/bazo_miner/protocol"
	"github.com/lisgie/bazo_miner/storage"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

//Tests block adding, verification, serialization and deserialization
//This test goes further than protocol/block_test.go because it tests the integrity of the payloads as well
//while protocol/block_test.go only tests serialization/deserialization and size calculation
func TestBlock(t *testing.T) {

	cleanAndPrepare()

	b := newBlock([32]byte{})
	hashFundsSlice, hashAccSlice, hashConfigSlice := createBlockWithTxs(b)
	finalizeBlock(b)

	encodedBlock := b.Encode()
	var decodedBlock *protocol.Block
	decodedBlock = decodedBlock.Decode(encodedBlock)
	err := validateBlock(decodedBlock)

	b.StateCopy = nil
	decodedBlock.StateCopy = nil

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
		t.Error("ConfigTx data is not properly serialized!")
	}
	if !reflect.DeepEqual(b, decodedBlock) {
		t.Error("Either serialization or deserialization failed, blocks are not equal!")
	}
}

//Duplicate Txs are not allowed
func TestBlockTxDuplicates(t *testing.T) {

	cleanAndPrepare()
	b := newBlock([32]byte{})
	createBlockWithTxs(b)

	finalizeBlock(b)

	//This is a normal block validation, should pass
	if err := validateBlock(b); err != nil {
		t.Errorf("Block validation failed.\n")
	}

	//Rollback the block and add a duplicate
	validateBlockRollback(b)
	if len(b.ConfigTxData) > 0 {
		b.ConfigTxData = append(b.ConfigTxData, b.ConfigTxData[0])
	}

	finalizeBlock(b)

	if err := validateBlock(b); err == nil {
		t.Errorf("Duplicate Tx not detected.\n")
	}
}

//Blocks that link to the previous block and have valid txs should pass
func TestMultipleBlocks(t *testing.T) {

	cleanAndPrepare()
	b := newBlock([32]byte{})
	createBlockWithTxs(b)
	finalizeBlock(b)
	if err := validateBlock(b); err != nil {
		t.Errorf("Block validation for (%v) failed: %v\n", b, err)
	}

	b2 := newBlock(b.Hash)
	createBlockWithTxs(b2)
	finalizeBlock(b2)
	if err := validateBlock(b2); err != nil {
		t.Errorf("Block validation failed: %v\n", err)
	}

	b3 := newBlock(b2.Hash)
	createBlockWithTxs(b3)
	finalizeBlock(b3)
	if err := validateBlock(b3); err != nil {
		t.Errorf("Block validation failed: %v\n", err)
	}

	b4 := newBlock(b3.Hash)
	createBlockWithTxs(b4)
	finalizeBlock(b4)
	if err := validateBlock(b4); err != nil {
		t.Errorf("Block validation failed: %v\n", err)
	}
}

//Test the blocktimestamp check
func TestTimestampCheck(t *testing.T) {

	cleanAndPrepare()
	timePast := time.Now().Unix() - 4000
	timeFuture := time.Now().Unix() + 4000
	timeNow := time.Now().Unix() + 50

	if err := timestampCheck(timePast); err == nil {
		t.Error("Dynamic time check failed\n")
	}

	if err := timestampCheck(timeFuture); err == nil {
		t.Error("Dynamic time check failed\n")
	}

	if err := timestampCheck(timeNow); err != nil {
		t.Errorf("Valid time got rejected: %v\n", err)
	}
}

//Helper function used by lots of test to fill the block with some random datas
func createBlockWithTxs(b *protocol.Block) ([][32]byte, [][32]byte, [][32]byte) {

	var testSize uint32
	testSize = 100

	var hashFundsSlice [][32]byte
	var hashAccSlice [][32]byte
	var hashConfigSlice [][32]byte
	//in order to create valid funds transactions we need to know the tx count of acc A

	rand := rand.New(rand.NewSource(time.Now().Unix()))
	loopMax := int(rand.Uint32()%testSize) + 1
	loopMax += int(accA.TxCnt)
	for cnt := int(accA.TxCnt); cnt < loopMax; cnt++ {
		accAHash := serializeHashContent(accA.Address)
		accBHash := serializeHashContent(accB.Address)
		tx, _ := protocol.ConstrFundsTx(0x01, rand.Uint64()%100+1, rand.Uint64()%100+1, uint32(cnt), accAHash, accBHash, &PrivKeyA)
		if err := addTx(b, tx); err == nil {
			//Might  be that we generated a block that was already generated before
			if storage.ReadOpenTx(tx.Hash()) != nil || storage.ReadClosedTx(tx.Hash()) != nil {
				continue
			}
			hashFundsSlice = append(hashFundsSlice, tx.Hash())
			storage.WriteOpenTx(tx)
		}
	}

	loopMax = int(rand.Uint32()%testSize) + 1
	for cnt := 0; cnt < loopMax; cnt++ {
		tx, _ := protocol.ConstrAccTx(0, rand.Uint64()%100+1, &RootPrivKey)
		if err := addTx(b, tx); err == nil {
			if storage.ReadOpenTx(tx.Hash()) != nil || storage.ReadClosedTx(tx.Hash()) != nil{
				continue
			}
			hashAccSlice = append(hashAccSlice, tx.Hash())
			storage.WriteOpenTx(tx)
		}
	}

	//NrConfigTx is saved in a uint8, so testsize shouldn't be larger than 255
	loopMax = int(rand.Uint32()%testSize) + 1
	for cnt := 0; cnt < loopMax; cnt++ {
		tx, _ := protocol.ConstrConfigTx(uint8(rand.Uint32()%256), uint8(rand.Uint32()%10+1), rand.Uint64()%2342873423, rand.Uint64()%1000+1, uint8(cnt), &RootPrivKey)

		if storage.ReadOpenTx(tx.Hash()) != nil || storage.ReadClosedTx(tx.Hash()) != nil {
			continue
		}

		//don't mess with the minimum fee and block size
		if tx.Id == 3 || tx.Id == 1 {
			continue
		}
		if err := addTx(b, tx); err == nil {

			hashConfigSlice = append(hashConfigSlice, tx.Hash())
			storage.WriteOpenTx(tx)
		}
	}

	return hashFundsSlice, hashAccSlice, hashConfigSlice
}