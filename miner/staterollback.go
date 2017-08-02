package miner

import (
	"github.com/lisgie/bazo_miner/protocol"
	"github.com/lisgie/bazo_miner/storage"
)

func accStateChangeRollback(txSlice []*protocol.AccTx) {

	for _, tx := range txSlice {
		accHash := serializeHashContent(tx.PubKey)

		acc := storage.State[accHash]
		if acc == nil {
			logger.Fatal("CRITICAL: An account that should have been saved does not exist!")
		}
		delete(storage.State, accHash)
	}
}

func fundsStateChangeRollback(txSlice []*protocol.FundsTx) {

	//Rollback in reverse order than original state change
	for cnt := len(txSlice) - 1; cnt >= 0; cnt-- {
		tx := txSlice[cnt]

		accSender, accReceiver := storage.GetAccountFromHash(tx.From), storage.GetAccountFromHash(tx.To)
		accSender.TxCnt -= 1
		accSender.Balance += tx.Amount
		accReceiver.Balance -= tx.Amount
	}
}

func configStateChangeRollback(txSlice []*protocol.ConfigTx, blockHash [32]byte) {

	if len(txSlice) == 0 {
		return
	}
	//Only rollback if the config changes lead to a parameterChange
	//there might be the case that the client is not running the latest version, it's still confirming
	//the transaction but does not understand the ID and thus is not changing the state
	if parameterSlice[len(parameterSlice)-1].blockHash != blockHash {
		return
	}
	//remove the latest entry in the parameters slice$
	parameterSlice = parameterSlice[:len(parameterSlice)-1]
	activeParameters = &parameterSlice[len(parameterSlice)-1]
}

func collectTxFeesRollback(accTx []*protocol.AccTx, fundsTx []*protocol.FundsTx, configTx []*protocol.ConfigTx, minerHash [32]byte) {

	minerAcc := storage.GetAccountFromHash(minerHash)
	//subtract fees from sender (check if that is allowed has already been done in the block validation)
	for _, tx := range accTx {
		//Money was created out of thin air, no need to write back
		minerAcc.Balance -= tx.Fee
	}

	for _, tx := range fundsTx {
		minerAcc.Balance -= tx.Fee
		senderAcc := storage.GetAccountFromHash(tx.From)
		senderAcc.Balance += tx.Fee
	}

	for _, tx := range configTx {
		//Money was created out of thin air, no need to write back
		minerAcc.Balance -= tx.Fee
	}
}

func collectBlockRewardRollback(reward uint64, minerHash [32]byte) {

	minerAcc := storage.GetAccountFromHash(minerHash)
	minerAcc.Balance -= reward
}
