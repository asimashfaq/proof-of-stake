package miner

import (
	"github.com/lisgie/bazo_miner/p2p"
	"github.com/lisgie/bazo_miner/protocol"
	"github.com/lisgie/bazo_miner/storage"
	"log"
)

func incomingData() {
	for {
		select {
		case tx := <-p2p.TxsIn:
			processTx(tx)
		case block := <-p2p.BlockIn:
			processBlock(block)
		}
	}
}

func processTx(incomingTx p2p.TxInfo) {

	var tx protocol.Transaction
	switch incomingTx.TxType {
	case p2p.FUNDSTX_BRDCST:
		var fTx *protocol.FundsTx
		fTx = fTx.Decode(incomingTx.Payload)

		if fTx == nil {
			return
		}
		tx = fTx
	case p2p.ACCTX_BRDCST:
		var aTx *protocol.AccTx
		aTx = aTx.Decode(incomingTx.Payload)
		if aTx == nil {
			return
		}
		tx = aTx
	case p2p.CONFIGTX_BRDCST:
		var cTx *protocol.ConfigTx
		cTx = cTx.Decode(incomingTx.Payload)
		if cTx == nil {
			return
		}
		tx = cTx
	}
	if storage.ReadOpenTx(tx.Hash()) != nil {
		return
	}
	if storage.ReadClosedTx(tx.Hash()) != nil {
		return
	}

	//write to mempool
	storage.WriteOpenTx(tx)
	p2p.TxsOut<-incomingTx
}

func processBlock(payload []byte) {

	var block *protocol.Block
	block = block.Decode(payload)

	//block already confirmed and validated
	if storage.ReadBlock(block.Hash) != nil {
		return
	}

	//claim a lock and start validating
	err := validateBlock(block)
	if err != nil {
		//no conflict, giving away for broadcast
		log.Printf("Received block could not be validated: %v\n", err)
	} else {
		broadcastBlock(block)
	}
}

func broadcastTx(tx protocol.Transaction) {
	switch tx.(type) {
	case *protocol.FundsTx:
		p2p.TxsOut <- p2p.TxInfo{p2p.FUNDSTX_BRDCST, tx.Encode()}
	case *protocol.AccTx:
		p2p.TxsOut <- p2p.TxInfo{p2p.ACCTX_BRDCST, tx.Encode()}
	case *protocol.ConfigTx:
		p2p.TxsOut <- p2p.TxInfo{p2p.CONFIGTX_BRDCST, tx.Encode()}
	}
}

func broadcastBlock(block *protocol.Block) { p2p.BlockOut <- block.Encode() }
