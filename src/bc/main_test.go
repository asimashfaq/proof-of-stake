package bc

import (
	"crypto/ecdsa"
	"testing"
	"os"
	"math/big"
	"crypto/elliptic"
	"golang.org/x/crypto/sha3"
	"log"
	"io/ioutil"
)

var accA, accB Account
var PrivKeyA ecdsa.PrivateKey
var PubKeyA ecdsa.PublicKey
var RootPrivKey ecdsa.PrivateKey


func addTestingAccounts() {
	puba1,_ := new(big.Int).SetString(pubA1,16)
	puba2,_ := new(big.Int).SetString(pubA2,16)
	priva,_ := new(big.Int).SetString(privA,16)
	PubKeyA = ecdsa.PublicKey{
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

	accA = Account{Balance: 12345678}
	copy(accA.Address[0:32], PrivKeyA.PublicKey.X.Bytes())
	copy(accA.Address[32:64], PrivKeyA.PublicKey.Y.Bytes())
	accA.Hash = sha3.Sum256(accA.Address[:])

	//This one is just for testing purposes
	accB = Account{Balance: 87654321}
	copy(accB.Address[0:32], PrivKeyB.PublicKey.X.Bytes())
	copy(accB.Address[32:64], PrivKeyB.PublicKey.Y.Bytes())
	accB.Hash = sha3.Sum256(accB.Address[:])

	//just to bootstrap
	var shortHashA [8]byte
	var shortHashB [8]byte
	copy(shortHashA[:], accA.Hash[0:8])
	copy(shortHashB[:], accB.Hash[0:8])

	State[shortHashA] = append(State[shortHashA],&accA)
	State[shortHashB] = append(State[shortHashB],&accB)
}

func addRootAccounts() {

	var pubKey [64]byte

	pub1,_ := new(big.Int).SetString(RootPub1,16)
	pub2,_ := new(big.Int).SetString(RootPub2,16)
	priv,_ := new(big.Int).SetString(RootPriv,16)
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

	rootHash := serializeHashContent(pubKey[:])

	var shortRootHash [8]byte
	copy(shortRootHash[:], rootHash[0:8])
	rootAcc := Account{Hash:rootHash, Address:pubKey}
	State[shortRootHash] = append(State[shortRootHash], &rootAcc)
	RootKeys[rootHash] = &rootAcc
}

func TestMain(m *testing.M) {

	//initialize states
	State = make(map[[8]byte][]*Account)
	RootKeys = make(map[[32]byte]*Account)

	addTestingAccounts()
	addRootAccounts()
	//we don't want logging msgs when testing, designated messages
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}
