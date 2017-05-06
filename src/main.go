package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"bc"
	"fmt"
)

func main() {

	state := make(map[[64]byte]bc.Account)
	privA, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	privB, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	if err != nil {
		return
	}
	accA := bc.Account{Balance: 15}
	var idA [64]byte
	copy(idA[0:32], privA.PublicKey.X.Bytes())
	copy(idA[32:64], privA.PublicKey.Y.Bytes())

	accB := bc.Account{Balance: 12}
	var idB [64]byte
	copy(idB[0:32], privB.PublicKey.X.Bytes())
	copy(idB[32:64], privB.PublicKey.Y.Bytes())

	state[idA] = accA
	state[idB] = accB

	b := bc.NewBlock([32]byte{}, state)

	tx, err := bc.ConstrTx(0, 2, idA, idB, privA)
	tx2, err := bc.ConstrTx(0, 3, idB, idA, privB)
	tx3, err := bc.ConstrTx(0, 1, idA, idB, privA)
	tx4, err := bc.ConstrTx(0, 4, idB, idA, privB)
	tx5, err := bc.ConstrTx(0, 3, idA, idB, privA)
	tx6, err := bc.ConstrTx(0, 1, idB, idA, privB)

	b.AddTx(&tx)
	b.AddTx(&tx2)
	b.AddTx(&tx3)
	b.AddTx(&tx4)
	b.AddTx(&tx5)
	b.AddTx(&tx6)

	if err != nil {
		return
	}

	b.FinalizeBlock()

	toSend := bc.EncodeForSend(tx)
	fmt.Printf("%x\n", toSend)
	toRcv := bc.DecodeForReceive(toSend)
	fmt.Printf("%x\n", toRcv.(bc.Transaction))

}

