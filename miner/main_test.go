package miner

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"github.com/lisgie/bazo_miner/p2p"
	"github.com/lisgie/bazo_miner/protocol"
	"github.com/lisgie/bazo_miner/storage"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"testing"
)

var PrivKeyA, PrivKeyB, MinerPrivKey ecdsa.PrivateKey
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
	hashA = serializeHashContent(accA.Address)

	//This one is just for testing purposes
	copy(accB.Address[0:32], PrivKeyB.PublicKey.X.Bytes())
	copy(accB.Address[32:64], PrivKeyB.PublicKey.Y.Bytes())
	hashB := serializeHashContent(accB.Address)

	//just to bootstrap
	storage.State[hashA] = accA
	storage.State[hashB] = accB

	minerPrivKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	var pubKey [64]byte
	var shortMiner [8]byte
	copy(pubKey[:32], minerPrivKey.X.Bytes())
	copy(pubKey[32:], minerPrivKey.Y.Bytes())
	minerHash := serializeHashContent(pubKey)
	copy(shortMiner[:], minerHash[0:8])
	minerAcc.Address = pubKey
	storage.State[minerHash] = minerAcc

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

	rootAcc := protocol.Account{Address: pubKey}
	storage.State[rootHash] = &rootAcc
	storage.RootKeys[rootHash] = &rootAcc
}

func cleanAndPrepare() {

	storage.DeleteAll()
	tmpState := make(map[[32]byte]*protocol.Account)
	tmpRootKeys := make(map[[32]byte]*protocol.Account)

	storage.State = tmpState
	storage.RootKeys = tmpRootKeys

	lastBlock = nil

	localBlockCount = 0
	globalBlockCount = 0
	genesis := newBlock([32]byte{})
	collectStatistics(genesis)
	storage.WriteClosedBlock(genesis)

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

	storage.Init("test.db")
	p2p.Init("8000")

	//setting a new random seed
	addTestingAccounts()
	addRootAccounts()
	//we don't want logging msgs when testing, designated messages
	logger = log.New(nil, "", 0)
	logger.SetOutput(ioutil.Discard)
	os.Exit(m.Run())

	storage.TearDown()
}
