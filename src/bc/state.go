package bc

import (
	"errors"
	"golang.org/x/crypto/sha3"
	"log"
)

//possibility of state change
//1) exchange funds from tx
//2) revert funds from previous tx
//3) this doesn't need a rollback, because digitally signed

func accStateChange(acctx *accTx) {

	addressHash := sha3.Sum256(acctx.PubKey[:])
	if _,exists := State[addressHash]; exists {
		log.Printf("Address already exists in the state: %x\n", addressHash[0:4])
		return
	}
	log.Printf("Added hash to state: %x\n", addressHash)
	State[addressHash] = Account{}
}

func fundsStateChange(tx *fundsTx) error {

	//rollback
	if _, exists := State[tx.Payload.From]; !exists {
		log.Printf("Sender does not exist in the State: %x\n", tx.Payload.From[0:4])
		return errors.New("Sender does not exist in the State.")
	}

	if _, exists := State[tx.Payload.To]; !exists {
		log.Printf("Receiver does not exist in the State: %x\n", tx.Payload.To[0:4])
		return errors.New("Receiver does not exist in the State.")
	}

	if tx.Payload.Amount > 0 {
		if uint64(tx.Payload.Amount) > State[tx.Payload.From].Balance {
			log.Printf("Sender does not have enough balance: %x\n", State[tx.Payload.From].Balance)
			return errors.New("Sender does not have enough funds for the transaction.")
		}
	}

	accSender := State[tx.Payload.From]
	accSender.TxCnt += 1
	accSender.Balance -= uint64(tx.Payload.Amount)
	State[tx.Payload.From] = accSender

	accReceiver := State[tx.Payload.To]
	accReceiver.Balance += uint64(tx.Payload.Amount)
	State[tx.Payload.To] = accReceiver

	PrintState()
	return nil
}

func fundsStateRollback(txSlice []fundsTx, index int) {

	for cnt := index; index >= 0; index-- {
		tx := txSlice[cnt]
		accSender := State[tx.Payload.From]
		accSender.TxCnt -= 1
		accSender.Balance += uint64(tx.Payload.Amount)
		State[tx.Payload.From] = accSender

		accReceiver := State[tx.Payload.To]
		accReceiver.Balance -= uint64(tx.Payload.Amount)
		State[tx.Payload.To] = accReceiver
	}
	PrintState()
}

func PrintState() {
	log.Println("State updated: ")
	for key,val := range State {
		log.Printf("%x: %v\n", key[0:4], val)
	}
}
