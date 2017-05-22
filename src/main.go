package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"bc"
	"golang.org/x/crypto/sha3"
	"time"
)


func listenForData() {
	//for now mock data
	//will be later exchanged with listening on a socket
	for i := 0; i < 100; i++ {
		err := bc.AddFundsTx(uint32(i), accA.Hash, accB.Hash, 3, privA)
		if err != nil {
			return
		}
		time.Sleep(time.Second)
	}
}

var accA, accB bc.Account
var privA *ecdsa.PrivateKey

func main() {

	bc.InitSystem()

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

	privA, _ = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	privB, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	if err != nil {
		return
	}

	//This is the client's account
	accA = bc.Account{Balance: 15000}
	copy(accA.Address[0:32], privA.PublicKey.X.Bytes())
	copy(accA.Address[32:64], privA.PublicKey.Y.Bytes())
	accA.Hash = sha3.Sum256(accA.Address[:])

	//This one is just for testing purposes
	accB = bc.Account{Balance: 702}
	copy(accB.Address[0:32], privB.PublicKey.X.Bytes())
	copy(accB.Address[32:64], privB.PublicKey.Y.Bytes())
	accB.Hash = sha3.Sum256(accB.Address[:])

	//just to bootstrap
	var shortHashA [8]byte
	var shortHashB [8]byte
	copy(shortHashA[:], accA.Hash[0:8])
	copy(shortHashB[:], accB.Hash[0:8])

	bc.State[shortHashA] = append(bc.State[shortHashA],&accA)
	bc.State[shortHashB] = append(bc.State[shortHashB],&accB)

	listenForData()

	bc.PrintState()
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

