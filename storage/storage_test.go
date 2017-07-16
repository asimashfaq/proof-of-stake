package storage

import (
	"github.com/lisgie/bazo_miner/protocol"
	"math/rand"
	"testing"
	"time"
)

//In-memory, k/v storage is tested with the test below
func TestReadWriteDeleteTx(t *testing.T) {

	rand := rand.New(rand.NewSource(time.Now().Unix()))

	accAHash := serializeHashContent(accA.Address)
	accBHash := serializeHashContent(accB.Address)

	var hashFundsSlice []*protocol.FundsTx
	var hashAccSlice []*protocol.AccTx
	var hashConfigSlice []*protocol.ConfigTx

	testsize := 1000

	loopMax := testsize
	for i := 0; i < loopMax; i++ {
		tx, _ := protocol.ConstrFundsTx(0x01, rand.Uint64()%100000+1, rand.Uint64()%10+1, uint32(i), accAHash, accBHash, &PrivKeyA)
		WriteOpenTx(tx)
		hashFundsSlice = append(hashFundsSlice, tx)
	}

	loopMax = testsize
	for i := 0; i < loopMax; i++ {
		tx, _ := protocol.ConstrAccTx(0, rand.Uint64()%100+1, &RootPrivKey)
		tx.Hash()
		WriteOpenTx(tx)
		hashAccSlice = append(hashAccSlice, tx)
	}

	loopMax = 256
	for cnt := 0; cnt < loopMax; cnt++ {
		tx, _ := protocol.ConstrConfigTx(uint8(rand.Uint32()%256), uint8(rand.Uint32()%5+1), rand.Uint64()%2342873423, rand.Uint64()%1000+1, &RootPrivKey)
		//don't mess with the minimum fee
		hashConfigSlice = append(hashConfigSlice, tx)
		WriteOpenTx(tx)
	}

	for _, tx := range hashFundsSlice {
		if ReadOpenTx(tx.Hash()) == nil {
			t.Errorf("Error writing transaction hash: %x\n", tx)
		}
	}

	for _, tx := range hashAccSlice {
		if ReadOpenTx(tx.Hash()) == nil {
			t.Errorf("Error writing transaction hash: %x\n", tx)
		}
	}

	for _, tx := range hashConfigSlice {
		if ReadOpenTx(tx.Hash()) == nil {
			t.Errorf("Error writing transaction hash: %x\n", tx)
		}
	}

	//deleting open txs
	for _, tx := range hashFundsSlice {
		DeleteOpenTx(tx)
	}

	for _, tx := range hashAccSlice {
		DeleteOpenTx(tx)
	}

	for _, tx := range hashConfigSlice {
		DeleteOpenTx(tx)
	}

	for _, tx := range hashFundsSlice {
		if ReadOpenTx(tx.Hash()) != nil {
			t.Errorf("Error deleting transaction hash: %x\n", tx)
		}
	}

	for _, tx := range hashAccSlice {
		if ReadOpenTx(tx.Hash()) != nil {
			t.Errorf("Error deleting transaction hash: %x\n", tx)
		}
	}

	for _, tx := range hashConfigSlice {
		if ReadOpenTx(tx.Hash()) != nil {
			t.Errorf("Error deleting transaction hash: %x\n", tx)
		}
	}
}

func TestReadWriteDeleteBlock(t *testing.T) {

	//this shouldn't panic
	DeleteOpenBlock([32]byte{'0'})

	b, b2, b3 := new(protocol.Block), new(protocol.Block), new(protocol.Block)
	b.Hash = [32]byte{'0'}
	b2.Hash = [32]byte{'1'}
	b3.Hash = [32]byte{'2'}
	WriteOpenBlock(b)
	WriteOpenBlock(b2)
	WriteOpenBlock(b3)

	if ReadOpenBlock(b.Hash) == nil || ReadOpenBlock(b2.Hash) == nil || ReadOpenBlock(b3.Hash) == nil {
		t.Error("Failed to write block to open block storage.")
	}

	newb1 := ReadOpenBlock(b.Hash)
	newb2 := ReadOpenBlock(b2.Hash)
	newb3 := ReadOpenBlock(b3.Hash)

	DeleteOpenBlock(newb1.Hash)
	DeleteOpenBlock(newb2.Hash)
	DeleteOpenBlock(newb3.Hash)

	WriteClosedBlock(newb1)
	WriteClosedBlock(newb2)
	WriteClosedBlock(newb3)

	if ReadOpenBlock(newb1.Hash) != nil ||
		ReadOpenBlock(newb2.Hash) != nil ||
		ReadOpenBlock(newb3.Hash) != nil ||
		ReadClosedBlock(b.Hash) == nil ||
		ReadClosedBlock(b2.Hash) == nil ||
		ReadClosedBlock(b3.Hash) == nil {
		t.Error("Failed to write block to kv storage.")
	}
}
