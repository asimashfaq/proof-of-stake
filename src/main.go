package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"bc"
	"golang.org/x/crypto/sha3"
)

func main() {


	bc.InitSystem()

	privA, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	privB, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	if err != nil {
		return
	}

	//This is the client's account
	accA := bc.Account{Balance: 15}
	copy(accA.Address[0:32], privA.PublicKey.X.Bytes())
	copy(accA.Address[32:64], privA.PublicKey.Y.Bytes())
	hashA := sha3.Sum256(accA.Address[:])


	//This one is just for testing purposes
	accB := bc.Account{Balance: 12}
	copy(accB.Address[0:32], privB.PublicKey.X.Bytes())
	copy(accB.Address[32:64], privB.PublicKey.Y.Bytes())
	hashB := sha3.Sum256(accB.Address[:])

	//just to bootstrap
	bc.State[hashA] = accA
	bc.State[hashB] = accB

	bc.PrintState()

	bc.AddFundsTx(0, hashA, hashB, 10, privA)
	bc.AddFundsTx(0, hashB, hashA, 3, privB)
	bc.AddFundsTx(1, hashA, hashB, 1, privA)
	bc.AddFundsTx(1, hashB, hashA, 4, privB)
	bc.AddFundsTx(2, hashA, hashB, 3, privA)
	bc.AddFundsTx(2, hashB, hashA, 2, privB)

	bc.FinalizeBlock()
	bc.ValidateBlock()

	/*toSend := bc.EncodeForSend(tx)
	fmt.Printf("%x\n", toSend)
	toRcv := bc.DecodeForReceive(toSend)
	fmt.Printf("%x\n", toRcv.(bc.fundsTx))*/

}

