package p2p

import (
	"github.com/lisgie/bazo_miner/protocol"
)

var (
	//Block from the network, to the miner
	BlockIn chan []byte = make(chan []byte)
	//Block from the miner, to the network
	BlockOut chan []byte = make(chan []byte)

	//Data requested by miner, to allow parallelism, we have a chan for every tx type
	FundsTxChan = make(chan *protocol.FundsTx)
	AccTxChan = make(chan *protocol.AccTx)
	ConfigTxChan = make(chan *protocol.ConfigTx)
	BlockReqChan = make(chan []byte)
)

//This is for blocks and txs that the miner successfully validated
func receiveBlockFromMiner() {
	for {
		block := <-BlockOut
		toBrdcst := BuildPacket(BLOCK_BRDCST, block)
		brdcstMsg <- toBrdcst
	}
}

func forwardBlockToMiner(p *peer, payload []byte) {
	BlockIn <- payload
}

//These are transactions the miner specifically requested
func forwardTxReqToMiner(p *peer, payload []byte, txType uint8) {
	if payload == nil {
		return
	}

	switch txType {
	case FUNDSTX_RES:
		var fundsTx *protocol.FundsTx
		fundsTx = fundsTx.Decode(payload)
		if fundsTx == nil {
			return
		}
		FundsTxChan <- fundsTx
	case ACCTX_RES:
		var accTx *protocol.AccTx
		accTx = accTx.Decode(payload)
		if accTx == nil {
			return
		}
		AccTxChan <- accTx
	case CONFIGTX_RES:
		var configTx *protocol.ConfigTx
		configTx = configTx.Decode(payload)
		if configTx == nil {
			return
		}
		ConfigTxChan <- configTx
	}
}

func forwardBlockReqToMiner(p *peer, payload []byte) {
	BlockReqChan <- payload
}

func ReadSystemTime() int64 {
	return systemTime
}
