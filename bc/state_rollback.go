package bc

func fundsStateChangeRollback(txSlice []*fundsTx) {

	for cnt := len(txSlice) - 1; cnt >= 0; cnt-- {
		tx := txSlice[cnt]

		accSender, accReceiver := getAccountFromHash(tx.fromHash), getAccountFromHash(tx.toHash)

		accSender.TxCnt -= 1
		accSender.Balance += tx.Amount

		accReceiver.Balance -= tx.Amount
	}
}

//this only happens for complete block rollbacks, therefore no index because everything has to be rolled back
func accStateChangeRollback(txSlice []*accTx) {

	for _, tx := range txSlice {
		accHash := serializeHashContent(tx.PubKey)

		var fixedHash [8]byte
		copy(fixedHash[:], accHash[0:8])

		accSlice := State[fixedHash]
		for i := range accSlice {
			accSliceHash := serializeHashContent(accSlice[i].Address)
			if accSliceHash == accHash {
				//deleting the account from the state
				//https://github.com/golang/go/wiki/SliceTricks, preventing mem leaks
				copy(accSlice[i:], accSlice[i+1:])
				accSlice[len(accSlice)-1] = nil
				accSlice = accSlice[:len(accSlice)-1]
			}
		}
		//preventing memory leaks, this is important
		if len(accSlice) == 0 {
			delete(State, fixedHash)
		}
	}
}

func configStateChangeRollback(txSlice []*configTx) {

	if len(txSlice) == 0 {
		return
	}
	//remove the latest entry in the parameters slice$
	parameterSlice = parameterSlice[:len(parameterSlice)-1]
	activeParameters = &parameterSlice[len(parameterSlice)-1]
}

func collectTxFeesRollback(fundsTx []*fundsTx, accTx []*accTx, configTx []*configTx, minerHash [32]byte) {

	miner := getAccountFromHash(minerHash)
	//subtract fees from sender (check if that is allowed has already been done in the block validation)
	for _, tx := range fundsTx {
		miner.Balance -= tx.Fee

		senderAcc := getAccountFromHash(tx.fromHash)
		senderAcc.Balance += tx.Fee
	}

	for _, tx := range accTx {
		//money gets created from thin air
		//no need to subtract money from root key
		miner.Balance -= tx.Fee
	}

	for _, tx := range configTx {
		//no need to subtract money from root key
		miner.Balance -= tx.Fee
	}
}

func collectBlockRewardRollback(reward uint64, minerHash [32]byte) {

	miner := getAccountFromHash(minerHash)
	miner.Balance -= reward
}
