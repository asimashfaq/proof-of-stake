package bc

import (
	"testing"
	"time"
	"math/rand"
)

func TestReadWriteTx(t *testing.T) {

	rand := rand.New(rand.NewSource(time.Now().Unix()))

	accAHash := serializeHashContent(accA.Address)
	accBHash := serializeHashContent(accB.Address)

	var hashFundsSlice [][32]byte
	var hashAccSlice [][32]byte

	loopMax := int(rand.Uint32()%10000)
	for i := 0; i < loopMax; i++ {
		tx, _ := ConstrFundsTx(0x01,rand.Uint64()%100000+1, rand.Uint64()%10+1, uint32(i), accAHash, accBHash, &PrivKeyA)
		writeOpenFundsTx(tx)
		hashFundsSlice = append(hashFundsSlice,hashFundsTx(tx))
	}

	loopMax = int(rand.Uint32()%10000)
	for i := 0; i < loopMax; i++ {
		tx,_:=ConstrAccTx(rand.Uint64()%100+1, &RootPrivKey)
		writeOpenAccTx(tx)
		hashAccSlice = append(hashAccSlice, hashAccTx(tx))
	}

	for _,hash := range hashFundsSlice {
		if readOpenFundsTx(hash) == nil {
			t.Errorf("Error writing transaction hash: %x\n", hash)
		}
	}

	for _,hash := range hashAccSlice {
		if readOpenAccTx(hash) == nil {
			t.Errorf("Error writing transaction hash: %x\n", hash)
		}
	}

	//deleting open txs
	for _,hash := range hashFundsSlice {
		deleteOpenFundsTx(hash)
	}

	for _,hash := range hashAccSlice {
		deleteOpenAccTx(hash)
	}

	for _,hash := range hashFundsSlice {
		if readOpenFundsTx(hash) != nil {
			t.Errorf("Error deleting transaction hash: %x\n", hash)
		}
	}

	for _,hash := range hashAccSlice {
		if readOpenAccTx(hash) != nil {
			t.Errorf("Error deleting transaction hash: %x\n", hash)
		}
	}
}

