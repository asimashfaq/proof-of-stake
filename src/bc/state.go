package bc

import (
	"log"
	"golang.org/x/crypto/sha3"
	"bytes"
	"errors"
	"encoding/binary"
)

func getAccountFromShortHash(hash [32]byte) (*Account) {

	var fixedHash [8]byte
	copy(fixedHash[:],hash[0:8])
	for _,acc := range State[fixedHash] {
		if bytes.Compare(acc.Hash[:],hash[:]) == 0 {
			return acc
		}
	}
	return nil
}

//possibility of state change
//1) exchange funds from tx
//2) revert funds from previous tx
//3) this doesn't need a rollback, because digitally signed
func accStateChange(acctx *accTx) {

	var fixedHash [8]byte
	addressHash := sha3.Sum256(acctx.PubKey[:])
	acc := getAccountFromShortHash(addressHash)
	if acc != nil {
		log.Printf("Address already exists in the state: %x\n", addressHash[0:4])
		return
	}
	copy(fixedHash[:],addressHash[0:8])
	newAcc := Account{Hash:addressHash}
	State[fixedHash] = append(State[fixedHash],&newAcc)
	PrintState()
}

func fundsStateChange(tx *fundsTx) error {

	accSender, accReceiver := getAccountFromShortHash(tx.fromHash), getAccountFromShortHash(tx.toHash)

	if accSender == nil {
		log.Printf("Sender does not exist in the State: %x\n", tx.fromHash[0:8])
		return errors.New("Sender does not exist in the State.")
	}

	if accReceiver == nil {
		log.Printf("Receiver does not exist in the State: %x\n", tx.toHash[0:8])
		return errors.New("Receiver does not exist in the State.")
	}

	amount := binary.BigEndian.Uint32(tx.Amount[:])
	if uint64(amount) > accSender.Balance {
		log.Printf("Sender does not have enough balance: %x\n", accSender.Balance)
		return errors.New("Sender does not have enough funds for the transaction.")
	}

	//we're manipulating pointer, no need to write back
	accSender.TxCnt += 1
	accSender.Balance -= uint64(amount)

	accReceiver.Balance += uint64(amount)

	PrintState()
	return nil
}

func fundsStateRollback(txSlice []fundsTx, index int) {

	/*for cnt := index; index >= 0; index-- {
		tx := txSlice[cnt]
		accSender := State[tx.Payload.From]
		accSender.TxCnt -= 1
		accSender.Balance += uint64(tx.Payload.Amount)
		State[tx.Payload.From] = accSender

		accReceiver := State[tx.Payload.To]
		accReceiver.Balance -= uint64(tx.Payload.Amount)
		State[tx.Payload.To] = accReceiver
	}
	PrintState()*/
}

func PrintState() {
	log.Println("State updated: ")
	for key,val := range State {
		for _,acc := range val {
			log.Printf("%x: %v\n", key[0:4], acc)
		}
	}
}
