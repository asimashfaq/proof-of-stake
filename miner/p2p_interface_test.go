package miner

import (
	"testing"
	"github.com/lisgie/bazo_miner/protocol"
	"time"
	"math/rand"
	"github.com/lisgie/bazo_miner/p2p"
	"fmt"
)

//mocking incoming transactions and blocks from outside

func TestIncomingData(t *testing.T) {

	cleanAndPrepare()
	go Init()
	var testSize uint32
	testSize = 100
	rand := rand.New(rand.NewSource(time.Now().Unix()))
	loopMax := int(rand.Uint32()%testSize) + 1
	for cnt := 0; cnt < loopMax; cnt++ {
		accAHash := serializeHashContent(accA.Address)
		accBHash := serializeHashContent(accB.Address)
		tx, _ := protocol.ConstrFundsTx(0x01, rand.Uint64()%100+1, rand.Uint64()%100+1, uint32(cnt), accAHash, accBHash, &PrivKeyA)
		fmt.Printf("%v\n", p2p.TxInfo{p2p.FUNDSTX_BRDCST,tx.Encode()})
		time.Sleep(time.Second)
		p2p.TxsIn<-p2p.TxInfo{p2p.FUNDSTX_BRDCST, tx.Encode()}
	}

	for {
		if globalBlockCount >= 6 {
			break
			time.Sleep(1*time.Second)
		}
	}

}
