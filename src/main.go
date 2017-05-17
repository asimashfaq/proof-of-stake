package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"bc"
	"golang.org/x/crypto/sha3"
	"fmt"
)

var foo bc.Account


func somefunc() (*bc.Account) {
	return &foo
}


func main() {

	bc.InitSystem()

	foo.Balance = 5
	bar := somefunc()

	bar.Balance = 2
	fmt.Printf("%v, %v\n", foo, bar)

	privA, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	privB, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	if err != nil {
		return
	}

	//This is the client's account
	accA := bc.Account{Balance: 928}
	copy(accA.Address[0:32], privA.PublicKey.X.Bytes())
	copy(accA.Address[32:64], privA.PublicKey.Y.Bytes())
	accA.Hash = sha3.Sum256(accA.Address[:])


	//This one is just for testing purposes
	accB := bc.Account{Balance: 702}
	copy(accB.Address[0:32], privB.PublicKey.X.Bytes())
	copy(accB.Address[32:64], privB.PublicKey.Y.Bytes())
	accB.Hash = sha3.Sum256(accB.Address[:])

	//just to bootstrap
	var shortHashA [8]byte
	var shortHashB [8]byte
	copy(shortHashA[:], accA.Hash[0:8])
	copy(shortHashB[:], accB.Hash[0:8])

	bc.State[shortHashA] = append(bc.State[shortHashA],accA)
	bc.State[shortHashB] = append(bc.State[shortHashB],accB)

	bc.PrintState()

	bc.AddFundsTx(0, accA.Hash, accB.Hash, 10, privA)
	bc.AddFundsTx(0, accB.Hash, accA.Hash, 2, privB)
	bc.AddFundsTx(1, accA.Hash, accB.Hash, 1, privA)

	newAddr := bc.AddAccTx()
	newHash := sha3.Sum256(newAddr.PubKey[:])

	bc.AddFundsTx(1, accB.Hash, accA.Hash, 4, privB)
	bc.AddFundsTx(2, accA.Hash, accB.Hash, 3, privA)
	bc.AddFundsTx(2, accB.Hash, accA.Hash, 2, privB)

	bc.FinalizeBlock()
	bc.ValidateBlock()

	bc.AddFundsTx(3, accA.Hash, accB.Hash, 32, privA)
	bc.AddFundsTx(3, accB.Hash, accA.Hash, 64, privB)
	bc.AddFundsTx(4, accA.Hash, accB.Hash, 10000, privA)
	bc.AddFundsTx(4, accA.Hash, newHash, 1, privA)

	bc.AddAccTx()
	bc.AddAccTx()
	bc.AddAccTx()
	bc.AddAccTx()

	bc.FinalizeBlock()
	bc.ValidateBlock()


	/*toSend := bc.EncodeForSend(tx)
	fmt.Printf("%x\n", toSend)
	toRcv := bc.DecodeForReceive(toSend)
	fmt.Printf("%x\n", toRcv.(bc.fundsTx))*/
}

