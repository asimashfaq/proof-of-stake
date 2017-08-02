package miner

import (
	"errors"
	"github.com/lisgie/bazo_miner/protocol"
	"github.com/lisgie/bazo_miner/storage"
	"golang.org/x/crypto/sha3"
)

func isRootKey(hash [32]byte) bool {
	_, exists := storage.RootKeys[hash]
	return exists
}

//for normal accounts, it
func accStateChange(txSlice []*protocol.AccTx) error {

	for _, tx := range txSlice {
		switch tx.Header {
		case 1:
			//first bit set, given account will be a new root account
			newAcc := protocol.Account{Address: tx.PubKey}
			storage.RootKeys[sha3.Sum256(tx.PubKey[:])] = &newAcc
			continue
		case 2:
			//second bit set, delete account from root account
			delete(storage.RootKeys, sha3.Sum256(tx.PubKey[:]))
			continue
		}

		//create a regular account
		addressHash := sha3.Sum256(tx.PubKey[:])
		acc := storage.GetAccountFromHash(addressHash)
		if acc != nil {
			logger.Printf("CRITICAL: Address already exists in the state: %x\n", addressHash[0:4])
			return errors.New("CRITICAL: Address already exists in the state")
		}
		newAcc := protocol.Account{Address: tx.PubKey}
		storage.State[addressHash] = &newAcc
	}
	return nil
}

func fundsStateChange(txSlice []*protocol.FundsTx) error {

	for index, tx := range txSlice {

		var err error
		//check if we have to issue new coins
		for hash, rootAcc := range storage.RootKeys {
			if hash == tx.From {
				logger.Printf("Root Key Transaction: %x\n", hash[0:8])

				if rootAcc.Balance+tx.Amount+tx.Fee > MAX_MONEY {
					logger.Printf("Root Account overflows (%v) with given transaction amount (%v) and fee (%v).\n", rootAcc.Balance, tx.Amount, tx.Fee)
					err = errors.New("Sender does not exist in the State.")
				}

				rootAcc.Balance += tx.Amount
				rootAcc.Balance += tx.Fee
			}
		}

		accSender, accReceiver := storage.GetAccountFromHash(tx.From), storage.GetAccountFromHash(tx.To)
		if accSender == nil {
			logger.Printf("CRITICAL: Sender does not exist in the State: %x\n", tx.From[0:8])
			err = errors.New("Sender does not exist in the State.")
		}

		if accReceiver == nil {
			logger.Printf("CRITICAL: Receiver does not exist in the State: %x\n", tx.To[0:8])
			err = errors.New("Receiver does not exist in the State.")
		}

		//also check for txCnt
		if tx.TxCnt != accSender.TxCnt {
			logger.Printf("Sender txCnt does not match: %v (tx.txCnt) vs. %v (state txCnt)\n", tx.TxCnt, accSender.TxCnt)
			err = errors.New("TxCnt mismatch!")
		}

		if (tx.Amount + tx.Fee) > accSender.Balance {
			logger.Printf("Sender does not have enough balance: %x\n", accSender.Balance)
			err = errors.New("Sender does not have enough funds for the transaction.")
		}

		//overflow protection
		if tx.Amount+accReceiver.Balance > MAX_MONEY {
			logger.Printf("Transaction amount (%v) would lead to balance overflow at the receiver account (%v)\n", tx.Amount, accReceiver.Balance)
			err = errors.New("Transaction amount would lead to balance overflow at the receiver account\n")
		}

		if err != nil {
			//was it the first tx in the block, no rollback needed
			if index == 0 {
				return err
			}
			fundsStateChangeRollback(txSlice[0 : index-1])
			return err
		}

		//we're manipulating pointer, no need to write back
		accSender.TxCnt += 1
		accSender.Balance -= tx.Amount
		accReceiver.Balance += tx.Amount
	}

	return nil
}

//we accept config slices with unknown id, but don't act on the payload
func configStateChange(configTxSlice []*protocol.ConfigTx, blockHash [32]byte) {

	var newParameters parameters
	//initialize it to state right now (before validating config txs)
	newParameters = *activeParameters

	if len(configTxSlice) == 0 {
		return
	}
	var change bool
	for _, tx := range configTxSlice {
		switch tx.Id {
		case protocol.FEE_MINIMUM_ID:
			if parameterBoundsChecking(protocol.FEE_MINIMUM_ID, tx.Payload) {
				newParameters.fee_minimum = tx.Payload
				//minor change, changes a runtime parameter, no further adaptations
				change = true
			}
		case protocol.BLOCK_SIZE_ID:
			if parameterBoundsChecking(protocol.BLOCK_SIZE_ID, tx.Payload) {
				newParameters.block_size = tx.Payload
				change = true
			}
		case protocol.BLOCK_REWARD_ID:
			if parameterBoundsChecking(protocol.BLOCK_REWARD_ID, tx.Payload) {
				newParameters.block_reward = tx.Payload
				change = true
			}

			//the following parameter changes all influence the timestamp process
			//we therefore need to reset the difficulty calculation
		case protocol.DIFF_INTERVAL_ID:
			if parameterBoundsChecking(protocol.DIFF_INTERVAL_ID, tx.Payload) {
				newParameters.diff_interval = tx.Payload
				change = true
			}
		case protocol.BLOCK_INTERVAL_ID:
			if parameterBoundsChecking(protocol.BLOCK_INTERVAL_ID, tx.Payload) {
				newParameters.block_interval = tx.Payload
				change = true
			}
		}
	}

	//only add a new parameter struct if something meaningful actually changed
	if change {
		newParameters.blockHash = blockHash

		parameterSlice = append(parameterSlice, newParameters)
		activeParameters = &parameterSlice[len(parameterSlice)-1]
	}

	//some parameters require more changes than just updating a runtime variable

}

func collectTxFees(accTxSlice []*protocol.AccTx, fundsTxSlice []*protocol.FundsTx, configTxSlice []*protocol.ConfigTx, minerHash [32]byte) error {

	var tmpAccTx []*protocol.AccTx
	var tmpFundsTx []*protocol.FundsTx
	var tmpConfigTx []*protocol.ConfigTx

	minerAcc := storage.GetAccountFromHash(minerHash)

	for _, tx := range accTxSlice {
		if minerAcc.Balance+tx.Fee > MAX_MONEY {
			//rollback of all perviously transferred transaction fees to the protocol's account
			collectTxFeesRollback(tmpAccTx, tmpFundsTx, tmpConfigTx, minerHash)
			logger.Printf("Miner balance (%v) overflows with transaction fee (%v).\n", minerAcc.Balance, tx.Fee)
			return errors.New("Miner balance overflows with transaction fee.\n")
		}

		//money gets created from thin air
		//no need to subtract money from root key
		minerAcc.Balance += tx.Fee
		tmpAccTx = append(tmpAccTx, tx)
	}

	//subtract fees from sender (check if that is allowed has already been done in the block validation)
	for _, tx := range fundsTxSlice {
		//preventing protocol account from overflowing
		if minerAcc.Balance+tx.Fee > MAX_MONEY {
			//rollback of all perviously transferred transaction fees to the protocol's account
			collectTxFeesRollback(tmpAccTx, tmpFundsTx, tmpConfigTx, minerHash)
			logger.Printf("Miner balance (%v) overflows with transaction fee (%v).\n", minerAcc.Balance, tx.Fee)
			return errors.New("Miner balance overflows with transaction fee.\n")
		}
		minerAcc.Balance += tx.Fee

		senderAcc := storage.GetAccountFromHash(tx.From)
		senderAcc.Balance -= tx.Fee

		tmpFundsTx = append(tmpFundsTx, tx)
	}

	for _, tx := range configTxSlice {
		if minerAcc.Balance+tx.Fee > MAX_MONEY {
			//rollback of all perviously transferred transaction fees to the protocol's account
			collectTxFeesRollback(tmpAccTx, tmpFundsTx, tmpConfigTx, minerHash)
			logger.Printf("Miner balance (%v) overflows with transaction fee (%v).\n", minerAcc.Balance, tx.Fee)
			return errors.New("Miner balance overflows with transaction fee.\n")
		}
		minerAcc.Balance += tx.Fee
		tmpConfigTx = append(tmpConfigTx, tx)
	}

	return nil
}

func collectBlockReward(reward uint64, minerHash [32]byte) error {
	miner := storage.GetAccountFromHash(minerHash)

	if miner == nil {
		return errors.New("Miner doesn't exist in the state!")
	}

	if miner.Balance+reward > MAX_MONEY {
		logger.Printf("Miner balance (%v) overflows with block reward (%v).\n", miner.Balance, reward)
		return errors.New("Miner balance overflows with transaction fee.\n")
	}
	miner.Balance += reward
	return nil
}

func printState() {
	logger.Println("State updated: ")
	for key, acc := range storage.State {
		logger.Printf("%x: %v\n", key[0:10], acc)
	}
}
