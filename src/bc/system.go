package bc

import (
	"crypto/ecdsa"
	"fmt"
	"log"
	"os"
	"time"
	"bytes"
	"encoding/binary"
)

//will act as interface to bc package
var State map[[8]byte][]*Account
var LogFile *os.File
var block *Block

func InitSystem() {

	LogFile, _ = os.OpenFile("log "+time.Now().String(), os.O_RDWR | os.O_CREATE , 0666)
	log.SetOutput(LogFile)

	log.Println("Starting system, initializing state map")
	State = make(map[[8]byte][]*Account)
	//temporary
	block = newBlock([32]byte{})
	//this is the responsibility of the client to send the right txCnt
}

func AddFundsTx(localTxCnt uint64, from, to [32]byte, amount uint32, key *ecdsa.PrivateKey) error {
	var header byte
	//constrFundsTx(header, amount [4]byte, txCnt [3]byte, from, to [32]byte, key *ecdsa.PrivateKey) (tx fundsTx, err error) {
	var buf bytes.Buffer
	var amountBuf [4]byte
	binary.Write(&buf, binary.BigEndian, amount)

	copy(amountBuf[:],buf.Bytes())
	buf.Reset()
	var txCntBuf [3]byte
	binary.Write(&buf,binary.BigEndian, localTxCnt)
	copy(txCntBuf[:],buf.Bytes())
	tx,err := constrFundsTx(header, amountBuf,txCntBuf, from,to, key)
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

func AddAccTx() *accTx {

	tx,err := constrAccTx()

	if err != nil {
		log.Printf("%v\n", err)
		//return errors.New("Failed to construct account tx.")
	}
	block.addTx(&tx)
	//return nil
	return &tx
}

//temporary
func FinalizeBlock() {
	block.finalizeBlock()
}

func ValidateBlock() {

	if validateBlock(block) != nil {
		return
	}
	prevBlock := block
	block = newBlock(prevBlock.Hash)
}