package p2p

//All incoming messages are processed here and acted upon accordingly
func processIncomingMsg(p *peer, header *Header, payload []byte) {

	switch header.TypeID {
		//BROADCASTING
	case FUNDSTX_BRDCST:
		processTxBrdcst(p, payload, FUNDSTX_BRDCST)
	case ACCTX_BRDCST:
		processTxBrdcst(p, payload, ACCTX_BRDCST)
	case CONFIGTX_BRDCST:
		processTxBrdcst(p, payload, CONFIGTX_BRDCST)
	case BLOCK_BRDCST:
		forwardBlockToMiner(p, payload)
	case TIME_BRDCST:
		processTimeRes(p, payload)

		//Miner Requests
	case FUNDSTX_REQ:
		txRes(p, payload, FUNDSTX_REQ)
	case ACCTX_REQ:
		txRes(p, payload, ACCTX_REQ)
	case CONFIGTX_REQ:
		txRes(p, payload, CONFIGTX_REQ)
	case BLOCK_REQ:
		blockRes(p, payload)
	case ACC_REQ:
		accRes(p, payload)
	case MINER_PING:
		pongRes(p, payload)
	case NEIGHBOR_REQ:
		neighborRes(p)

		//Miner Responses
	case NEIGHBOR_RES:
		processNeighborRes(p, payload)
	case BLOCK_RES:
		forwardBlockReqToMiner(p, payload)
	case FUNDSTX_RES:
		forwardTxReqToMiner(p, payload, FUNDSTX_RES)
	case ACCTX_RES:
		forwardTxReqToMiner(p, payload, ACCTX_RES)
	case CONFIGTX_RES:
		forwardTxReqToMiner(p, payload, CONFIGTX_RES)
	}
}