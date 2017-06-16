package bc

import (
	"errors"
	"fmt"
	"golang.org/x/crypto/sha3"
	"log"
)

func isRootKey(hash [32]byte) bool {

	_, exists := RootKeys[hash]
	return exists
}

func getAccountFromHash(hash [32]byte) *Account {

	var fixedHash [8]byte
	copy(fixedHash[:], hash[0:8])
	for _, acc := range State[fixedHash] {
		accHash := serializeHashContent(acc.Address)
		if accHash == hash {
			return acc
		}
	}
	return nil
}

func fundsStateChange(txSlice []*fundsTx) error {

	for index, tx := range txSlice {

		var err error
		//check if we have to issue new coins
		for hash, rootAcc := range RootKeys {
			if hash == tx.fromHash {
				log.Printf("Root Key Transaction: %x\n", hash[0:8])

				if rootAcc.Balance+tx.Amount+tx.Fee > MAX_MONEY {
					log.Printf("Root Account overflows (%v) with given transaction amount (%v) and fee (%v).\n", rootAcc.Balance, tx.Amount, tx.Fee)
					err = errors.New("Sender does not exist in the State.")
				}

				rootAcc.Balance += tx.Amount
				rootAcc.Balance += tx.Fee
			}
		}

		accSender, accReceiver := getAccountFromHash(tx.fromHash), getAccountFromHash(tx.toHash)
		if accSender == nil {
			log.Printf("CRITICAL: Sender does not exist in the State: %x\n", tx.fromHash[0:8])
			err = errors.New("Sender does not exist in the State.")
		}

		if accReceiver == nil {
			log.Printf("CRITICAL: Receiver does not exist in the State: %x\n", tx.toHash[0:8])
			err = errors.New("Receiver does not exist in the State.")
		}

		//also check for txCnt
		if tx.TxCnt != accSender.TxCnt {
			log.Printf("Sender txCnt does not match: %v (tx.txCnt) vs. %v (state txCnt)\n", tx.TxCnt, accSender.TxCnt)
			err = errors.New("TxCnt mismatch!")
		}

		if (tx.Amount + tx.Fee) > accSender.Balance {
			log.Printf("Sender does not have enough balance: %x\n", accSender.Balance)
			err = errors.New("Sender does not have enough funds for the transaction.")
		}

		//overflow protection
		if tx.Amount+accReceiver.Balance > MAX_MONEY {
			log.Printf("Transaction amount (%v) would lead to balance overflow at the receiver account (%v)\n", tx.Amount, accReceiver.Balance)
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

//possibility of state change
//1) exchange funds from tx
//2) revert funds from previous tx
//3) this doesn't need a rollback, because digitally signed
//https://golang.org/doc/faq#stack_or_heap
func accStateChange(txSlice []*accTx) error {

	for _, tx := range txSlice {
		var fixedHash [8]byte
		addressHash := sha3.Sum256(tx.PubKey[:])
		acc := getAccountFromHash(addressHash)
		if acc != nil {
			log.Printf("CRITICAL: Address already exists in the state: %x\n", addressHash[0:4])
			return errors.New("CRITICAL: Address already exists in the state")
		}
		copy(fixedHash[:], addressHash[0:8])
		newAcc := Account{Address: tx.PubKey}
		State[fixedHash] = append(State[fixedHash], &newAcc)
	}
	return nil
}

func configStateChange(configTxSlice []*configTx, blockHash [32]byte) {

	if len(configTxSlice) == 0 {
		return
	}

	for _, tx := range configTxSlice {
		switch tx.Id {
		case FEE_MINIMUM_ID:
			FEE_MINIMUM = tx.Payload
		case BLOCK_SIZE_ID:
			if tx.Payload == 0 {
				fmt.Printf("¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬%v\n", tx)
			}
			BLOCK_SIZE = tx.Payload
		case DIFF_INTERVAL_ID:
			DIFF_INTERVAL = tx.Payload
		case BLOCK_INTERVAL_ID:
			BLOCK_INTERVAL = tx.Payload
		case BLOCK_REWARD_ID:
			BLOCK_REWARD = tx.Payload
		}
	}
	parameterSlice = append(parameterSlice, parameters{
		blockHash,
		FEE_MINIMUM,
		BLOCK_SIZE,
		DIFF_INTERVAL,
		BLOCK_INTERVAL,
		BLOCK_REWARD,
	})
	activeParameters = &parameterSlice[len(parameterSlice)-1]
}

func collectTxFees(fundsTxSlice []*fundsTx, accTxSlice []*accTx, configTxSlice []*configTx, minerHash [32]byte) error {

	var tmpFundsTx []*fundsTx
	var tmpAccTx []*accTx
	var tmpConfigTx []*configTx

	miner := getAccountFromHash(minerHash)

	//subtract fees from sender (check if that is allowed has already been done in the block validation)
	for _, tx := range fundsTxSlice {
		//preventing miner account from overflowing
		if miner.Balance+tx.Fee > MAX_MONEY {
			//rollback of all perviously transferred transaction fees to the miner's account
			collectTxFeesRollback(tmpFundsTx, tmpAccTx, tmpConfigTx, minerHash)
			log.Printf("Miner balance (%v) overflows with transaction fee (%v).\n", miner.Balance, tx.Fee)
			return errors.New("Miner balance overflows with transaction fee.\n")
		}
		miner.Balance += tx.Fee

		senderAcc := getAccountFromHash(tx.fromHash)
		senderAcc.Balance -= tx.Fee

		tmpFundsTx = append(tmpFundsTx, tx)
	}

	for _, tx := range accTxSlice {
		if miner.Balance+tx.Fee > MAX_MONEY {
			//rollback of all perviously transferred transaction fees to the miner's account
			collectTxFeesRollback(tmpFundsTx, tmpAccTx, tmpConfigTx, minerHash)
			log.Printf("Miner balance (%v) overflows with transaction fee (%v).\n", miner.Balance, tx.Fee)
			return errors.New("Miner balance overflows with transaction fee.\n")
		}

		//money gets created from thin air
		//no need to subtract money from root key
		miner.Balance += tx.Fee
		tmpAccTx = append(tmpAccTx, tx)
	}

	for _, tx := range configTxSlice {
		if miner.Balance+tx.Fee > MAX_MONEY {
			//rollback of all perviously transferred transaction fees to the miner's account
			collectTxFeesRollback(tmpFundsTx, tmpAccTx, tmpConfigTx, minerHash)
			log.Printf("Miner balance (%v) overflows with transaction fee (%v).\n", miner.Balance, tx.Fee)
			return errors.New("Miner balance overflows with transaction fee.\n")
		}
		miner.Balance += tx.Fee
		tmpConfigTx = append(tmpConfigTx, tx)
	}

	return nil
}

func collectBlockReward(reward uint64, minerHash [32]byte) error {
	miner := getAccountFromHash(minerHash)

	if miner.Balance+reward > MAX_MONEY {
		log.Printf("Miner balance (%v) overflows with block reward (%v).\n", miner.Balance, reward)
		return errors.New("Miner balance overflows with transaction fee.\n")
	}
	miner.Balance += reward
	return nil
}

func PrintState() {
	log.Println("State updated: ")
	for key, val := range State {
		for _, acc := range val {
			log.Printf("%x: %v\n", key[0:8], acc)
		}
	}
}
