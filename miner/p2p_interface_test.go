
package miner

import (
"testing"
"math/rand"
"time"
"github.com/lisgie/bazo_miner/protocol"
"github.com/lisgie/bazo_miner/p2p"
)

var done chan struct{}

//mocking incoming transactions and blocks from outside
func TestIncomingData(t *testing.T) {

	cleanAndPrepare()
	go Init()

	done = make(chan struct{})

	go fundstx()
	go acctx()
	go configtx()

	for cnt := 0; cnt < 3; cnt++ {
		<-done
	}

	tmpCount := globalBlockCount
	//at this point all transactions have been written
	//wait two more blocks to make sure all transactions have been validated
	for {
		time.Sleep(1*time.Second)
		if globalBlockCount >= tmpCount+2 {
			break
		}
	}
}

func fundstx() {
	rand := rand.New(rand.NewSource(time.Now().Unix()))
	for cnt := 0; cnt < 2; cnt++ {
		accAHash := serializeHashContent(accA.Address)
		accBHash := serializeHashContent(accB.Address)
		tx, _ := protocol.ConstrFundsTx(0x01, rand.Uint64()%100+1, rand.Uint64()%100+1, uint32(cnt), accAHash, accBHash, &PrivKeyA)
		time.Sleep(10*time.Second)
		p2p.TxsIn<-p2p.TxInfo{p2p.FUNDSTX_BRDCST, tx.Encode()}
	}
	done<-struct{}{}
}

func acctx() {
	rand := rand.New(rand.NewSource(time.Now().Unix()))
	for cnt := 0; cnt < 10; cnt++ {
		tx, err := protocol.ConstrAccTx(0, rand.Uint64()%100+1, &RootPrivKey)
		if err == nil {
			p2p.TxsIn<-p2p.TxInfo{p2p.ACCTX_BRDCST,tx.Encode()}
			time.Sleep(time.Second)
		}
	}
	done<-struct{}{}
}

func configtx() {
	rand := rand.New(rand.NewSource(time.Now().Unix()))
	for cnt := 0; cnt < 10; cnt++ {
		tx, err := protocol.ConstrConfigTx(0, uint8(rand.Uint64()%10)+1, rand.Uint64()%12328738, rand.Uint64()%10000, &RootPrivKey)
		//don't mess with the fee interval
		if err == nil && tx.Id != 3{
			p2p.TxsIn<-p2p.TxInfo{p2p.CONFIGTX_BRDCST,tx.Encode()}
			time.Sleep(time.Second)
		}
	}
	done<-struct{}{}
}



