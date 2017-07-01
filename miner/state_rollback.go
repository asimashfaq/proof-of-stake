package miner

import (
	"github.com/lisgie/bazo_miner/protocol"
	"github.com/lisgie/bazo_miner/storage"
	"log"
)

func fundsStateChangeRollback(txSlice []*protocol.FundsTx) {

	for cnt := len(txSlice) - 1; cnt >= 0; cnt-- {
		tx := txSlice[cnt]

		accSender, accReceiver := storage.GetAccountFromHash(tx.From), storage.GetAccountFromHash(tx.To)

		accSender.TxCnt -= 1
		accSender.Balance += tx.Amount

		accReceiver.Balance -= tx.Amount
	}
}

//this only happens for complete block rollbacks, therefore no index because everything has to be rolled back
func accStateChangeRollback(txSlice []*protocol.AccTx) {

	for _, tx := range txSlice {
		accHash := serializeHashContent(tx.PubKey)

		acc := storage.State[accHash]
		if acc == nil {
			log.Fatal("An account that should have been saved does not exist!")
		}
		delete(storage.State, accHash)
	}
}

func configStateChangeRollback(txSlice []*protocol.ConfigTx, blockHash [32]byte) {

	if len(txSlice) == 0 {
		return
	}
	//only rollback if the config changes lead to a parameterChange
	//there might be the case that the client is not running the latest version, it's still confirming
	//the transaction but does not understand the ID and thus is not changing the state
	if parameterSlice[len(parameterSlice)-1].blockHash != blockHash {
		return
	}
	//remove the latest entry in the parameters slice$
	parameterSlice = parameterSlice[:len(parameterSlice)-1]
	activeParameters = &parameterSlice[len(parameterSlice)-1]
}

func collectTxFeesRollback(fundsTx []*protocol.FundsTx, accTx []*protocol.AccTx, configTx []*protocol.ConfigTx, minerHash [32]byte) {

	minerAcc := storage.GetAccountFromHash(minerHash)
	//subtract fees from sender (check if that is allowed has already been done in the block validation)
	for _, tx := range fundsTx {
		minerAcc.Balance -= tx.Fee

		senderAcc := storage.GetAccountFromHash(tx.From)
		senderAcc.Balance += tx.Fee
	}

	for _, tx := range accTx {
		//money gets created from thin air
		//no need to subtract money from root key
		minerAcc.Balance -= tx.Fee
	}

	for _, tx := range configTx {
		//no need to subtract money from root key
		minerAcc.Balance -= tx.Fee
	}
}

func collectBlockRewardRollback(reward uint64, minerHash [32]byte) {

	minerAcc := storage.GetAccountFromHash(minerHash)
	minerAcc.Balance -= reward
}
