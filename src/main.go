package main

import (
	//"fmt"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"bc"
)


func main() {

	privA, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	privB, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	if err != nil {
		return
	}

	var accA bc.Account
	accA.Balance = 10
	copy(accA.Id[0:32], privA.PublicKey.X.Bytes())
	copy(accA.Id[32:64], privA.PublicKey.Y.Bytes())

	var accB bc.Account
	copy(accB.Id[0:32], privB.PublicKey.X.Bytes())
	copy(accB.Id[32:64], privB.PublicKey.Y.Bytes())

	bc.ConstrTransact(0, 2, accA, accB, privA)

	//fmt.Printf("%x\n", tx)


}

