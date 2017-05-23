package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"bc"
	"golang.org/x/crypto/sha3"
	"net"
	"fmt"
	"bufio"
	"math/big"
)

const (
	pubA1 = "c2be9abbeaec39a066c2a09cee23bb9ab2a0b88f2880b1e785b4d317adf0dc7c"
	pubA2 = "8ce020fde838d9c443f6c93345dafe7fd74f091c4d2f30b37e2453679a257ed5"
	privA = "ba127fa8f802b008b9cdb58f4e44809d48f1b000cff750dda9cd6b312395c1c5"
	pubB1 = "5d7eefd58e3d2f309471928ab4bbd104e52973372c159fa652b8ca6b57ff68b8"
	pubB2 = "ab301a6a77b201c416ddc13a2d33fdf200a5302f6f687e0ea09085debaf8a1d9"
	privB = "7a0a9babcc97ea7991ed67ed7f800f70c5e04e99718960ad8efab2ca052f00c7"
)

func listenForData() {
	//for now mock data
	//will be later exchanged with listening on a socket

	ln, _ := net.Listen("tcp", ":8081")
	conn, _ := ln.Accept()
	var a []byte
	a = make([]byte, 100)

	for {
		reader := bufio.NewReader(conn)

		len,_ := reader.ReadByte()
		fmt.Printf("%v\n", len)
		//for {
		reader.Read(a)

		fmt.Printf("%x\n", a[1:len])
		reader.Reset(conn)
	}

}


func setUpTestSystem() {

	var accA, accB bc.Account

	puba1,_ := new(big.Int).SetString(pubA1,16)
	puba2,_ := new(big.Int).SetString(pubA2,16)
	priva,_ := new(big.Int).SetString(privA,16)
	PubKeyA := ecdsa.PublicKey{
		elliptic.P256(),
		puba1,
		puba2,
	}
	PrivKeyA := ecdsa.PrivateKey{
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

	//just to bootstrap
	var shortHashA [8]byte
	var shortHashB [8]byte
	copy(shortHashA[:], accA.Hash[0:8])
	copy(shortHashB[:], accB.Hash[0:8])

	bc.State[shortHashA] = append(bc.State[shortHashA],&accA)
	bc.State[shortHashB] = append(bc.State[shortHashB],&accB)
}

func main() {

	bc.InitSystem()
	bc.Sync()
	setUpTestSystem()

	listenForData()

	//open networking connection


	/*_rootPub1,_ := new(big.Int).SetString(bc.RootPub1,16)
	_rootPub2,_ := new(big.Int).SetString(bc.RootPub2,16)
	_rootPriv,_ := new(big.Int).SetString(bc.RootPriv,16)
	rootPubKey := ecdsa.PublicKey{
		elliptic.P256(),
		_rootPub1,
		_rootPub2,
	}
	rootPrivKey := ecdsa.PrivateKey{
		rootPubKey,
		_rootPriv,
	}
	var rootPubKeyHash [32]byte
	var tmp [64]byte
	copy(tmp[:32],_rootPub1.Bytes())
	copy(tmp[32:],_rootPub2.Bytes())
	rootPubKeyHash = sha3.Sum256(tmp[:])*/




	//listenForData()

	/*bc.AddFundsTx(0, rootPubKeyHash, accA.Hash, 100, &rootPrivKey)
	bc.AddFundsTx(0, accA.Hash, accB.Hash, 10, privA)

	bc.AddFundsTx(1, accA.Hash, accB.Hash, 500, privA)
	bc.AddFundsTx(2, accA.Hash, accB.Hash, 200, privA)
	bc.AddFundsTx(0, accB.Hash, accA.Hash, 2, privB)
	bc.AddFundsTx(3, accA.Hash, accB.Hash, 1, privA)

	newAddr := bc.AddAccTx()
	newHash := sha3.Sum256(newAddr.PubKey[:])

	bc.AddFundsTx(1, accB.Hash, accA.Hash, 4, privB)
	bc.AddFundsTx(4, accA.Hash, accB.Hash, 3, privA)
	bc.AddFundsTx(2, accB.Hash, accA.Hash, 2, privB)

	bc.FinalizeBlock()
	bc.ValidateBlock()

	bc.AddFundsTx(5, accA.Hash, accB.Hash, 32, privA)
	bc.AddFundsTx(3, accB.Hash, accA.Hash, 64, privB)
	bc.AddFundsTx(6, accA.Hash, accB.Hash, 10000, privA)
	bc.AddFundsTx(6, accA.Hash, newHash, 1, privA)

	bc.AddAccTx()
	bc.AddAccTx()
	bc.AddAccTx()
	bc.AddAccTx()

	bc.FinalizeBlock()
	bc.ValidateBlock()*/


	/*toSend := bc.EncodeForSend(tx)
	fmt.Printf("%x\n", toSend)
	toRcv := bc.DecodeForReceive(toSend)
	fmt.Printf("%x\n", toRcv.(bc.fundsTx))*/
}

