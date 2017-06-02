package main

import (
	"math/big"
	"crypto/ecdsa"
	"crypto/elliptic"
	"bc"
	"golang.org/x/crypto/sha3"
	"net"
	"time"
	"math/rand"
	"network"
	"bytes"
	"encoding/binary"
	"bufio"
	"fmt"
)

const(
	pubA1 = "c2be9abbeaec39a066c2a09cee23bb9ab2a0b88f2880b1e785b4d317adf0dc7c"
	pubA2 = "8ce020fde838d9c443f6c93345dafe7fd74f091c4d2f30b37e2453679a257ed5"
	privA = "ba127fa8f802b008b9cdb58f4e44809d48f1b000cff750dda9cd6b312395c1c5"
	pubB1 = "5d7eefd58e3d2f309471928ab4bbd104e52973372c159fa652b8ca6b57ff68b8"
	pubB2 = "ab301a6a77b201c416ddc13a2d33fdf200a5302f6f687e0ea09085debaf8a1d9"
	privB = "7a0a9babcc97ea7991ed67ed7f800f70c5e04e99718960ad8efab2ca052f00c7"
)

const (
	HEADER_LEN = 2
)

var accA, accB bc.Account
var PrivKeyA ecdsa.PrivateKey
var RootPrivKey ecdsa.PrivateKey

func main() {

	var header byte
	var txCnt uint32

	header = 0x02

	prepAccs()

	for {
		var conn net.Conn
		conn, _ = net.Dial("tcp", "127.0.0.1:8081")
		tx2,_ := bc.ConstrAccTx(rand.Uint64()%100+1,&RootPrivKey)
		accData := bc.EncodeAccTx(tx2)
		toSend2 := make([]byte, len(accData)+HEADER_LEN)
		toSend2[0] = byte(len(accData))
		toSend2[1] = network.ACCTX
		copy(toSend2[2:],accData)
		conn.Write(toSend2)

		conn.Close()
		time.Sleep(1*time.Second)

		conn, _ = net.Dial("tcp", "127.0.0.1:8081")
		tx, _ := bc.ConstrFundsTx(header,rand.Uint64()%100+1, rand.Uint64()%50+1, txCnt, accA.Hash,accB.Hash, &PrivKeyA)
		txCnt++
		fundsData := bc.EncodeFundsTx(tx)
		toSend := make([]byte, len(fundsData)+HEADER_LEN)
		toSend[0] = byte(len(fundsData))
		toSend[1] = network.FUNDSTX
		copy(toSend[2:],fundsData)
		conn.Write(toSend)

		time.Sleep(time.Second)

		conn.Close()

	}
}


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


	var pubKey [64]byte

	pub1,_ := new(big.Int).SetString(bc.RootPub1,16)
	pub2,_ := new(big.Int).SetString(bc.RootPub2,16)
	priv,_ := new(big.Int).SetString(bc.RootPriv,16)
	PubKeyA = ecdsa.PublicKey{
		elliptic.P256(),
		pub1,
		pub2,
	}
	RootPrivKey = ecdsa.PrivateKey{
		PubKeyA,
		priv,
	}

	copy(pubKey[32-len(pub1.Bytes()):32],pub1.Bytes())
	copy(pubKey[64-len(pub2.Bytes()):],pub2.Bytes())


	var buf bytes.Buffer

	binary.Write(&buf,binary.LittleEndian, pubKey[:])

	rootHash :=  sha3.Sum256(buf.Bytes())

	var shortRootHash [8]byte
	copy(shortRootHash[:], rootHash[0:8])
	//rootAcc := bc.Account{Hash:rootHash, Address:pubKey}
}