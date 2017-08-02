package miner

import (
	"github.com/lisgie/bazo_miner/p2p"
	"github.com/lisgie/bazo_miner/protocol"
	"github.com/lisgie/bazo_miner/storage"
)

//The code in this source file communicates with the p2p package via channels

//Constantly listen to incoming data from the network
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

	//Make sure the transaction can be properly decoded, verification is done at a later stage to reduce latency
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
		logger.Printf("Received transaction (%x) already in the mempool.\n", tx.Hash())
		return
	}
	if storage.ReadClosedTx(tx.Hash()) != nil {
		logger.Printf("Received transaction (%x) already validated.\n", tx.Hash())
		return
	}

	//Write to mempool and rebroadcast
	logger.Printf("Writing transaction (%x) in the mempool.\n", tx.Hash())
	storage.WriteOpenTx(tx)
	p2p.TxsOut <- incomingTx
}

func processBlock(payload []byte) {

	var block *protocol.Block
	block = block.Decode(payload)

	//Block already confirmed and validated
	if storage.ReadClosedBlock(block.Hash) != nil {
		logger.Printf("Received block (%x) has already been validated.\n", block.Hash[0:12])
		return
	}

	//Start validation process
	err := validateBlock(block)
	if err != nil {
		logger.Printf("Received block (%x) could not be validated: %v\n", block.Hash[0:12], err)
	} else {
		broadcastBlock(block)
	}
}

//p2p.BlockOut is a channel whose data get consumed by the p2p package
func broadcastBlock(block *protocol.Block) { p2p.BlockOut <- block.Encode() }