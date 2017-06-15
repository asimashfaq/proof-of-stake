package bc

import (
	"crypto/ecdsa"
	"testing"
	"os"
	"math/big"
	"crypto/elliptic"
	"crypto/rand"
	"io/ioutil"
	"log"
	"github.com/lisgie/bazo_miner/storage"
)

var accA, accB, minerAcc *Account
var PrivKeyA, PrivKeyB ecdsa.PrivateKey
var PubKeyA, PubKeyB ecdsa.PublicKey
var RootPrivKey ecdsa.PrivateKey

func addTestingAccounts() {

	accA,accB,minerAcc = new(Account),new(Account),new(Account)

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

	State[shortHashA] = append(State[shortHashA],accA)
	State[shortHashB] = append(State[shortHashB],accB)

	MinerPrivKey, _ = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	var pubKey [64]byte
	var shortMiner [8]byte
	copy(pubKey[:32],MinerPrivKey.X.Bytes())
	copy(pubKey[32:],MinerPrivKey.Y.Bytes())
	MinerHash = serializeHashContent(pubKey)
	copy(shortMiner[:],MinerHash[0:8])
	minerAcc.Address = pubKey
	State[shortMiner] = append(State[shortMiner],minerAcc)

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

	rootHash := serializeHashContent(pubKey)

	var shortRootHash [8]byte
	copy(shortRootHash[:], rootHash[0:8])
	rootAcc := Account{Address:pubKey}
	State[shortRootHash] = append(State[shortRootHash], &rootAcc)
	RootKeys[rootHash] = &rootAcc
}

func cleanAndPrepare() {

	deleteEverything()

	tmpState := make(map[[8]byte][]*Account)
	tmpRootKeys := make(map[[32]byte]*Account)

	State = tmpState
	RootKeys = tmpRootKeys

	genesis := newBlock()
	lastBlock = genesis
	writeBlock(genesis)

	localBlockCount = 0
	globalBlockCount = 0

	var tmpSlice []parameters
	var tmpTimestamp []int64

	timestamp = tmpTimestamp

	tmpSlice = append(tmpSlice,parameters{
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

	//initialize states
	State = make(map[[8]byte][]*Account)
	RootKeys = make(map[[32]byte]*Account)


	//set system parameters
	parameterSlice = append(parameterSlice,parameters{
		[32]byte{},
		1,
		1000,
		2016,
		60,
		0,
	})
	activeParameters = &parameterSlice[0]

	storage.Init()

	//genesis block
	genesis := newBlock()
	writeBlock(genesis)
	collectStatistics(genesis)

	//setting a new random seed
	addTestingAccounts()
	addRootAccounts()
	//we don't want logging msgs when testing, designated messages
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}
