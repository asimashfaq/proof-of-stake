package main

import (
	"encoding/binary"
	"bytes"
	"math/big"
	"crypto/ecdsa"
	"crypto/elliptic"
	"bc"
	"golang.org/x/crypto/sha3"
	"time"
	"net"
)

const(
	pubA1 = "c2be9abbeaec39a066c2a09cee23bb9ab2a0b88f2880b1e785b4d317adf0dc7c"
	pubA2 = "8ce020fde838d9c443f6c93345dafe7fd74f091c4d2f30b37e2453679a257ed5"
	privA = "ba127fa8f802b008b9cdb58f4e44809d48f1b000cff750dda9cd6b312395c1c5"
	pubB1 = "5d7eefd58e3d2f309471928ab4bbd104e52973372c159fa652b8ca6b57ff68b8"
	pubB2 = "ab301a6a77b201c416ddc13a2d33fdf200a5302f6f687e0ea09085debaf8a1d9"
	privB = "7a0a9babcc97ea7991ed67ed7f800f70c5e04e99718960ad8efab2ca052f00c7"
)

func prepAccs() {

	puba1,_ := new(big.Int).SetString(pubA1,16)
	puba2,_ := new(big.Int).SetString(pubA2,16)
	priva,_ := new(big.Int).SetString(privA,16)
	PubKeyA := ecdsa.PublicKey{
		elliptic.P256(),
		puba1,
		puba2,
	}
	PrivKeyA = ecdsa.PrivateKey{
		PubKeyA,
		priva,
	}

	//fmt.Printf("%x\n", puba1)

	pubb1,_ := new(big.Int).SetString(pubB1,16)
	pubb2,_ := new(big.Int).SetString(pubB2,16)
	privb,_ := new(big.Int).SetString(privB,16)
	PubKeyB := ecdsa.PublicKey{
		elliptic.P256(),
		pubb1,
		pubb2,
	}
	PrivKeyB := ecdsa.PrivateKey{
		PubKeyB,
		privb,
	}

	accA = bc.Account{Balance: 15000}
	copy(accA.Address[0:32], PrivKeyA.PublicKey.X.Bytes())
	copy(accA.Address[32:64], PrivKeyA.PublicKey.Y.Bytes())
	accA.Hash = sha3.Sum256(accA.Address[:])

	//This one is just for testing purposes
	accB = bc.Account{Balance: 702}
	copy(accB.Address[0:32], PrivKeyB.PublicKey.X.Bytes())
	copy(accB.Address[32:64], PrivKeyB.PublicKey.Y.Bytes())
	accB.Hash = sha3.Sum256(accB.Address[:])
}

var accA, accB bc.Account
var PrivKeyA ecdsa.PrivateKey

func main() {

	var buf bytes.Buffer
	var header byte
	var amountBuf [4]byte
	var txCntBuf [3]byte
	var feeBuf [2]byte
	var fee uint16
	var amount uint32
	var localTxCnt uint32
	fee = 1
	amount = 10
	localTxCnt = 0

	prepAccs()

	//this has to be easier
	var tmpTxCntBuf [4]byte


	binary.Write(&buf, binary.BigEndian, fee)
	copy(feeBuf[:],buf.Bytes())
	buf.Reset()

	binary.Write(&buf, binary.BigEndian, amount)
	copy(amountBuf[:],buf.Bytes())
	buf.Reset()

	conn, _ := net.Dial("tcp", "127.0.0.1:8081")

	for i := 0; i < 100; i++ {

		binary.Write(&buf, binary.BigEndian, localTxCnt)
		copy(tmpTxCntBuf[:],buf.Bytes())
		copy(txCntBuf[:],tmpTxCntBuf[1:])
		buf.Reset()
		localTxCnt+=1

		tx, err := bc.ConstrFundsTx(header, amountBuf, feeBuf, txCntBuf, accA.Hash,accB.Hash, &PrivKeyA)
		data := bc.EncodeFundsTx(tx)
		toSend := make([]byte, len(data)+1)
		toSend[0] = byte(len(data))
		copy(toSend[1:],data)
		conn.Write(toSend)

		//err := bc.AddFundsTx(uint32(i), accA.Hash, accB.Hash, 3, privA)
		if err != nil {
			return
		}

		time.Sleep(time.Second)
	}
}