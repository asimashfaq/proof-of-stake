package bc

import (
	"crypto/ecdsa"
	"fmt"
	"bytes"
	"encoding/gob"
)

//will act as interface to bc package
var State map[[32]byte]Account
var block *Block

func InitSystem() {


	foo := accTx{Sig:[64]byte{'1'}}
	var tx transaction
	tx = &foo

	var rcv transaction
	var buf bytes.Buffer
	//var tx transaction
	enc := gob.NewEncoder(&buf)
	enc.Encode(tx)
	dec := gob.NewDecoder(&buf)
	fmt.Printf("%x\n", buf.Bytes())
	gob.Register()
	dec.Decode(&rcv)
	fmt.Printf("%T\n", rcv)






	State = make(map[[32]byte]Account)
	//temporary
	block = newBlock([32]byte{})
	//this is the responsibility of the client to send the right txCnt
}

func AddAcc(hash [32]byte, acc Account) {
	State[hash] = acc
}

func AddFundsTx(localTxCnt uint64, from, to [32]byte, amount uint32, key *ecdsa.PrivateKey) error {
	tx, err := constrFundsTx(localTxCnt, amount, from, to, key)
	//localTxCnt++
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	err = block.addTx(&tx)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	return nil
}

func AddAccTx() error {

	tx,err := constrAccTx()
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	block.addTx(&tx)
	return nil
}

//temporary
func FinalizeBlock() {
	block.finalizeBlock()
	fmt.Printf("%x\n", block)
}

func ValidateBlock() {

	fmt.Printf("%v\n", validateBlock(block))
	fmt.Printf("%x\n", State)
}