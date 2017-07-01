package miner

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"
	"github.com/lisgie/bazo_miner/protocol"
	"github.com/lisgie/bazo_miner/storage"
	"log"
	"math/big"
	"os"
	"sync"
	"time"
)

var LogFile *os.File

//using these accounts a mining beneficiary
var accA, accB protocol.Account
var hashA, hashB [32]byte

var blockValidation = &sync.Mutex{}

var timestamp []int64
var parameterSlice []parameters
var activeParameters *parameters
var tmpSlice []parameters

func Sync() {

}

func Init() {

	testing_setup()

	LogFile, _ = os.OpenFile("log/miner "+time.Now().String(), os.O_RDWR|os.O_CREATE, 0666)
	log.SetOutput(LogFile)

	//var tmpTimestamp []int64
	parameterSlice = append(parameterSlice, parameters{
		[32]byte{},
		1,
		1000,
		2016,
		60,
		0,
	})
	activeParameters = &parameterSlice[0]

	localBlockCount = 0
	globalBlockCount = 0
	genesis := newBlock([32]byte{})
	collectStatistics(genesis)
	storage.WriteBlock(genesis)

	log.Println("Starting system, initializing state map")
	//genesisBlock := newBlock([32]byte{})
	//collectStatistics(genesisBlock)
	//storage.WriteBlock(genesisBlock)

	go incomingData()
	mining()
}

func mining() {
	currentBlock := newBlock([32]byte{})
	for {
		err := finalizeBlock(currentBlock)
		if err != nil {
			fmt.Printf("Mining failure: %v\n", err)
		} else {
			fmt.Println("Block mined.")
		}
		//else a block was received meanwhile that was added to the chain, all the effort was in vain :(
		//wait for lock here only
		if err != nil {
			log.Printf("%v\n", err)
		} else {
			broadcastBlock(currentBlock)
			validateBlock(currentBlock)
		}

		//TODO: Mutex for state validation, build new block to mine only AFTER state update (opentxs->closedtxs)
		//mining successful, construct new block out of mempool transactions
		nextBlock := newBlock(lastBlock.Hash)
		currentBlock = nextBlock
		prepareBlock(currentBlock)
	}
}

func prepareBlock(block *protocol.Block) {

	//empty mempool (opentxs)
	opentxs := storage.ReadAllOpenTxs()
	for _, tx := range opentxs {
		err := addTx(block, tx)
		if err != nil {
			storage.DeleteOpenTx(tx)
		}
	}
}

//some testing code
func testing_setup() {

	var pubKey [64]byte

	pub1, _ := new(big.Int).SetString(protocol.RootPub1, 16)
	pub2, _ := new(big.Int).SetString(protocol.RootPub2, 16)

	copy(pubKey[:32], pub1.Bytes())
	copy(pubKey[32:], pub2.Bytes())

	rootHash := serializeHashContent(pubKey)

	rootAcc := protocol.Account{Address: pubKey}
	storage.State[rootHash] = &rootAcc
	storage.RootKeys[rootHash] = &rootAcc

	puba1, _ := new(big.Int).SetString(protocol.PubA1, 16)
	puba2, _ := new(big.Int).SetString(protocol.PubA2, 16)
	priva, _ := new(big.Int).SetString(protocol.PrivA, 16)
	PubKeyA := ecdsa.PublicKey{
		elliptic.P256(),
		puba1,
		puba2,
	}
	PrivKeyA := ecdsa.PrivateKey{
		PubKeyA,
		priva,
	}

	pubb1, _ := new(big.Int).SetString(protocol.PubB1, 16)
	pubb2, _ := new(big.Int).SetString(protocol.PubB2, 16)
	privb, _ := new(big.Int).SetString(protocol.PrivB, 16)
	PubKeyB := ecdsa.PublicKey{
		elliptic.P256(),
		pubb1,
		pubb2,
	}
	PrivKeyB := ecdsa.PrivateKey{
		PubKeyB,
		privb,
	}

	accA = protocol.Account{Balance: 1500000}
	copy(accA.Address[0:32], PrivKeyA.PublicKey.X.Bytes())
	copy(accA.Address[32:64], PrivKeyA.PublicKey.Y.Bytes())
	hashA = serializeHashContent(accA.Address)

	//This one is just for testing purposes
	accB = protocol.Account{Balance: 702000}
	copy(accB.Address[0:32], PrivKeyB.PublicKey.X.Bytes())
	copy(accB.Address[32:64], PrivKeyB.PublicKey.Y.Bytes())
	hashB = serializeHashContent(accB.Address)

	storage.State[hashA] = &accA
	storage.State[hashB] = &accB
}
