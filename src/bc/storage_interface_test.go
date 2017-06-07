package bc

import (
	"testing"
	"time"
	"math/rand"
)

func TestReadWriteTx(t *testing.T) {

	rand := rand.New(rand.NewSource(time.Now().Unix()))

	var hashFundsSlice [][32]byte
	var hashAccSlice [][32]byte

	loopMax := int(rand.Uint32()%10000)
	for i := 0; i < loopMax; i++ {
		tx, _ := ConstrFundsTx(0x01,rand.Uint64()%100000+1, rand.Uint64()%10+1, uint32(i), accA.Hash, accB.Hash, &PrivKeyA)
		writeFundsTx(tx)
		hashFundsSlice = append(hashFundsSlice,hashFundsTx(tx))
	}

	loopMax = int(rand.Uint32()%10000)
	for i := 0; i < loopMax; i++ {
		tx,_:=ConstrAccTx(rand.Uint64()%100+1, &RootPrivKey)
		writeAccTx(tx)
		hashAccSlice = append(hashAccSlice, hashAccTx(tx))
	}


	for _,hash := range hashFundsSlice {
		if readFundsTx(hash) == nil {
			t.Errorf("Error writing transaction hash: %x\n", hash)
		}
	}

	for _,hash := range hashAccSlice {
		if readAccTx(hash) == nil {
			t.Errorf("Error writing transaction hash: %x\n", hash)
		}
	}
}

