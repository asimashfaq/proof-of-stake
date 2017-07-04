package p2p

import "fmt"

var (
	//Data from the network, for the miner
	TxsIn   chan TxInfo
	BlockIn chan []byte

	//Data from the miner, for the network
	TxsOut   chan TxInfo
	BlockOut chan []byte

	//Data requested by miner, for tx we don't need the type because we know what we request
	TxReqChan chan []byte
 	BlockReqChan chan []byte
)

//this is for blocks and txs that the miner successfully validated
func receiveDataFromMiner() {
	for {
		select {
		case block := <-BlockOut:
			logger.Printf("Received a block from the miner for broadcasting.")
			toBrdcst := BuildPacket(BLOCK_BRDCST, block)
			brdcstMsg <- toBrdcst
		case txInfo := <-TxsOut:
			logger.Printf("Received a transaction from the miner for broadcasting: ID: %v.\n", txInfo.TxType)
			toBrdcst := BuildPacket(txInfo.TxType, txInfo.Payload)
			brdcstMsg <- toBrdcst
		}
	}
}

//we can't broadcast incoming messages directly, need to forward them to the miner (to check if
//the tx has already been broadcast before, whether it was a valid tx at all)
func forwardTxToMiner(p *peer, payload []byte, brdcstType uint8) {
	logger.Printf("Received a transaction (ID: %v) from %v.\n", brdcstType, p.conn.RemoteAddr().String())
	TxsIn <- TxInfo{brdcstType, payload}
}
func forwardBlockToMiner(p *peer, payload []byte) {
	logger.Printf("Received a block from %v.\n", p.conn.RemoteAddr().String())
	BlockIn <- payload
}

//these are transactions the miner specifically requested
func forwardTxReqToMiner(p *peer, payload []byte) {
	logger.Printf("Received the response to a transaction request from %v.\n", p.conn.RemoteAddr().String())
	TxReqChan<-payload
}

func forwardBlockReqToMiner(p *peer, payload []byte) {
	fmt.Printf("Received the response to a block request from %v: %x.\n", p.conn.RemoteAddr().String(), payload)
	logger.Printf("Received the response to a block request from %v.\n", p.conn.RemoteAddr().String())
	BlockReqChan<-payload
}