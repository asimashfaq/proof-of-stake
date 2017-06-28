package miner

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"github.com/lisgie/bazo_miner/protocol"
	"github.com/lisgie/bazo_miner/storage"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"testing"
	"github.com/lisgie/bazo_miner/p2p"
)

var accA, accB, minerAcc *protocol.Account
var PrivKeyA, PrivKeyB ecdsa.PrivateKey
var PubKeyA, PubKeyB ecdsa.PublicKey
var RootPrivKey ecdsa.PrivateKey

func addTestingAccounts() {

	accA, accB, minerAcc = new(protocol.Account), new(protocol.Account), new(protocol.Account)

	puba1, _ := new(big.Int).SetString(protocol.PubA1, 16)
	puba2, _ := new(big.Int).SetString(protocol.PubA2, 16)
	priva, _ := new(big.Int).SetString(protocol.PrivA, 16)
	PubKeyA = ecdsa.PublicKey{
		elliptic.P256(),
		puba1,
		puba2,
	}
	PrivKeyA = ecdsa.PrivateKey{
		PubKeyA,
		priva,
	}

	pubb1, _ := new(big.Int).SetString(protocol.PubB1, 16)
	pubb2, _ := new(big.Int).SetString(protocol.PubB2, 16)
	privb, _ := new(big.Int).SetString(protocol.PrivB, 16)
	PubKeyB = ecdsa.PublicKey{
		elliptic.P256(),
		pubb1,
		pubb2,
	}
	PrivKeyB = ecdsa.PrivateKey{
		PubKeyB,
		privb,
	}

	copy(accA.Address[0:32], PrivKeyA.PublicKey.X.Bytes())
	copy(accA.Address[32:64], PrivKeyA.PublicKey.Y.Bytes())
	accAHash := serializeHashContent(accA.Address)

	//This one is just for testing purposes
	copy(accB.Address[0:32], PrivKeyB.PublicKey.X.Bytes())
	copy(accB.Address[32:64], PrivKeyB.PublicKey.Y.Bytes())
	accBHash := serializeHashContent(accB.Address)

	//just to bootstrap
	var shortHashA [8]byte
	var shortHashB [8]byte
	copy(shortHashA[:], accAHash[0:8])
	copy(shortHashB[:], accBHash[0:8])

	storage.State[shortHashA] = append(storage.State[shortHashA], accA)
	storage.State[shortHashB] = append(storage.State[shortHashB], accB)

	MinerPrivKey, _ = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	var pubKey [64]byte
	var shortMiner [8]byte
	copy(pubKey[:32], MinerPrivKey.X.Bytes())
	copy(pubKey[32:], MinerPrivKey.Y.Bytes())
	MinerHash = serializeHashContent(pubKey)
	copy(shortMiner[:], MinerHash[0:8])
	minerAcc.Address = pubKey
	storage.State[shortMiner] = append(storage.State[shortMiner], minerAcc)

}

func addRootAccounts() {

	var pubKey [64]byte

	pub1, _ := new(big.Int).SetString(protocol.RootPub1, 16)
	pub2, _ := new(big.Int).SetString(protocol.RootPub2, 16)
	priv, _ := new(big.Int).SetString(protocol.RootPriv, 16)
	PubKeyA = ecdsa.PublicKey{
		elliptic.P256(),
		pub1,
		pub2,
	}
	RootPrivKey = ecdsa.PrivateKey{
		PubKeyA,
		priv,
	}

	copy(pubKey[32-len(pub1.Bytes()):32], pub1.Bytes())
	copy(pubKey[64-len(pub2.Bytes()):], pub2.Bytes())

	rootHash := serializeHashContent(pubKey)

	var shortRootHash [8]byte
	copy(shortRootHash[:], rootHash[0:8])
	rootAcc := protocol.Account{Address: pubKey}
	storage.State[shortRootHash] = append(storage.State[shortRootHash], &rootAcc)
	storage.RootKeys[rootHash] = &rootAcc
}

func cleanAndPrepare() {

	storage.DeleteAll()
	tmpState := make(map[[8]byte][]*protocol.Account)
	tmpRootKeys := make(map[[32]byte]*protocol.Account)

	storage.State = tmpState
	storage.RootKeys = tmpRootKeys

	localBlockCount = 0
	globalBlockCount = 0
	genesis := newBlock([32]byte{})
	collectStatistics(genesis)
	storage.WriteBlock(genesis)

	var tmpSlice []parameters
	var tmpTimestamp []int64

	timestamp = tmpTimestamp

	tmpSlice = append(tmpSlice, parameters{
		[32]byte{},
		1,
		1000,
		2016,
		60,
		0,
	})
	parameterSlice = tmpSlice
	activeParameters = &parameterSlice[0]

	addTestingAccounts()
	addRootAccounts()

	minerAcc.Balance = 0
	accA.Balance = 123232345678
	accB.Balance = 823237654321
	accA.TxCnt = 0
	accB.TxCnt = 0
}

func TestMain(m *testing.M) {

	storage.Init()
	p2p.Init()

	//setting a new random seed
	addTestingAccounts()
	addRootAccounts()
	//we don't want logging msgs when testing, designated messages
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())

	storage.TearDown()
}
