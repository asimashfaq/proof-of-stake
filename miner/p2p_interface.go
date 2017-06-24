package miner

import (
	"github.com/lisgie/bazo_miner/protocol"
	"time"
)

//we need to decode incoming transactions, therefore type is needed
//for outgoing transactions, the p2p package needs the information to build the proper header
type txInfo struct {
	txType uint8
	payload []byte
}

//receiving txs, blocks etc. and giving free for broadcasting asnychronously
var (
	txsIn chan txInfo
	blockIn chan []byte

	txsOut chan txInfo
	blockOut chan []byte
)

func consumeTx() {

	for {
		if txQueue.Size() != 0 {
			nextBlockAccess.Lock()
			addTx(nextBlock, txQueue.Dequeue().(protocol.Transaction))
			nextBlockAccess.Unlock()
		}
		time.Sleep(20 * time.Millisecond)
	}
}