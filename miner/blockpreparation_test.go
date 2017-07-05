package miner

import (
	"testing"
	"fmt"
	"github.com/lisgie/bazo_miner/protocol"
	"math/rand"
	"github.com/lisgie/bazo_miner/storage"
	"time"
)

func TestPrepareAndSortTxs(t *testing.T) {

	cleanAndPrepare()

	//fill the open storage with fundstx
	rand := rand.New(rand.NewSource(time.Now().Unix()))
	for cnt := 0; cnt < 10; cnt++ {
		accAHash := serializeHashContent(accA.Address)
		accBHash := serializeHashContent(accB.Address)
		tx, _ := protocol.ConstrFundsTx(0x01, rand.Uint64()%100+1, rand.Uint64()%100+1, uint32(cnt), accAHash, accBHash, &PrivKeyA)
		if verifyFundsTx(tx) {
			storage.WriteOpenTx(tx)
		}
	}

	b := newBlock([32]byte{})
	prepareBlock(b)
	finalizeBlock(b)
	fmt.Printf("%v\n", b)
}
