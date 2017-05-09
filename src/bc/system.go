package bc

import (
	"crypto/ecdsa"
	"fmt"
)

//will act as interface to bc package
var State map[[32]byte]Account
var block *Block

func InitSystem() {
	State = make(map[[32]byte]Account)
	//temporary
	block = newBlock([32]byte{})
	//this is the responsibility of the client to send the right txCnt
}

func AddAcc(hash [32]byte, acc Account) {
	State[hash] = acc
}

func AddTx(localTxCnt uint64, from, to [32]byte, amount uint32, key *ecdsa.PrivateKey) error {
	tx, err := constrTx(localTxCnt, amount, from, to, key)
	//localTxCnt++
	if err != nil {
		return err
	}
	block.addTx(tx)
	return nil
}

//temporary
func FinalizeBlock() {
	block.finalizeBlock()
}

func ValidateBlock() {

	fmt.Printf("%v\n", validateBlock(block))
}