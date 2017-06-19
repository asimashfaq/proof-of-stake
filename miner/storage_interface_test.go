package miner

import (
	"github.com/lisgie/bazo_miner/protocol"
	"math/rand"
	"testing"
	"time"
)

func TestReadWriteTx(t *testing.T) {

	rand := rand.New(rand.NewSource(time.Now().Unix()))

	accAHash := serializeHashContent(accA.Address)
	accBHash := serializeHashContent(accB.Address)

	var hashFundsSlice [][32]byte
	var hashAccSlice [][32]byte
	var hashConfigSlice [][32]byte

	loopMax := int(rand.Uint32() % 100)
	for i := 0; i < loopMax; i++ {
		tx, _ := protocol.ConstrFundsTx(0x01, rand.Uint64()%100000+1, rand.Uint64()%10+1, uint32(i), accAHash, accBHash, &PrivKeyA)
		writeOpenTx(tx)
		hashFundsSlice = append(hashFundsSlice, tx.Hash())
	}

	loopMax = int(rand.Uint32() % 100)
	for i := 0; i < loopMax; i++ {
		tx, _ := protocol.ConstrAccTx(0, rand.Uint64()%100+1, &RootPrivKey)
		tx.Hash()
		writeOpenTx(tx)
		hashAccSlice = append(hashAccSlice, tx.Hash())
	}

	loopMax = int(rand.Uint32() % 100)
	for cnt := 0; cnt < loopMax; cnt++ {
		tx, _ := protocol.ConstrConfigTx(uint8(rand.Uint32()%256), uint8(rand.Uint32()%5+1), rand.Uint64()%2342873423, rand.Uint64()%1000+1, &RootPrivKey)
		//don't mess with the minimum fee
		hashConfigSlice = append(hashConfigSlice, tx.Hash())
		writeOpenTx(tx)
	}

	for _, hash := range hashFundsSlice {
		if readOpenFundsTx(hash) == nil {
			t.Errorf("Error writing transaction hash: %x\n", hash)
		}
	}

	for _, hash := range hashAccSlice {
		if readOpenAccTx(hash) == nil {
			t.Errorf("Error writing transaction hash: %x\n", hash)
		}
	}

	for _, hash := range hashConfigSlice {
		if readOpenConfigTx(hash) == nil {
			t.Errorf("Error writing transaction hash: %x\n", hash)
		}
	}

	//deleting open txs
	for _, hash := range hashFundsSlice {
		deleteOpenTx(hash)
	}

	for _, hash := range hashAccSlice {
		deleteOpenTx(hash)
	}

	for _, hash := range hashConfigSlice {
		deleteOpenTx(hash)
	}

	for _, hash := range hashFundsSlice {
		if readOpenFundsTx(hash) != nil {
			t.Errorf("Error deleting transaction hash: %x\n", hash)
		}
	}

	for _, hash := range hashAccSlice {
		if readOpenAccTx(hash) != nil {
			t.Errorf("Error deleting transaction hash: %x\n", hash)
		}
	}

	for _, hash := range hashConfigSlice {
		if readOpenConfigTx(hash) != nil {
			t.Errorf("Error deleting transaction hash: %x\n", hash)
		}
	}
}
