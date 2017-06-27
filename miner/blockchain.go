package miner

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
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
var currentBlock, nextBlock *protocol.Block

var MinerHash [32]byte
var MinerPrivKey *ecdsa.PrivateKey

var blockValidation = &sync.Mutex{}

var timestamp []int64
var parameterSlice []parameters
var activeParameters *parameters

func Sync() {

}

func Init() {

	testing_setup()

	LogFile, _ = os.OpenFile("logminer "+time.Now().String(), os.O_RDWR|os.O_CREATE, 0666)
	log.SetOutput(LogFile)

	log.Println("Starting system, initializing state map")
	genesisBlock := newBlock()
	collectStatistics(genesisBlock)
	storage.WriteBlock(genesisBlock)

	currentBlock = newBlock()
	nextBlock = newBlock()

	go incomingData()
	mining()
}

func mining() {
	for {
		err := finalizeBlock(currentBlock)

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
		prevHash := lastBlock.Hash
		currentBlock = newBlock()
		currentBlock.PrevHash = prevHash
		prepareBlock(currentBlock)
	}
}

func prepareBlock(block *protocol.Block) {

	//empty mempool (opentxs)
	opentxs := storage.ReadAllOpenTxs()
	for _,tx := range opentxs {
		addTx(block,tx)
	}
}

//some testing code
func testing_setup() {
	MinerPrivKey, _ = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	var pubKey [64]byte
	var shortMiner [8]byte
	copy(pubKey[:32], MinerPrivKey.X.Bytes())
	copy(pubKey[32:], MinerPrivKey.Y.Bytes())
	MinerHash = serializeHashContent(pubKey)
	copy(shortMiner[:], MinerHash[0:8])
	minerAcc := protocol.Account{Address: pubKey}
	storage.State[shortMiner] = append(storage.State[shortMiner], &minerAcc)

	pub1, _ := new(big.Int).SetString(protocol.RootPub1, 16)
	pub2, _ := new(big.Int).SetString(protocol.RootPub2, 16)

	copy(pubKey[:32], pub1.Bytes())
	copy(pubKey[32:], pub2.Bytes())

	rootHash := serializeHashContent(pubKey)

	var shortRootHash [8]byte
	copy(shortRootHash[:], rootHash[0:8])
	rootAcc := protocol.Account{Address: pubKey}
	storage.State[shortRootHash] = append(storage.State[shortRootHash], &rootAcc)
	storage.RootKeys[rootHash] = &rootAcc
	var accA, accB protocol.Account

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
	hashA := serializeHashContent(accA.Address)

	//This one is just for testing purposes
	accB = protocol.Account{Balance: 702000}
	copy(accB.Address[0:32], PrivKeyB.PublicKey.X.Bytes())
	copy(accB.Address[32:64], PrivKeyB.PublicKey.Y.Bytes())
	hashB := serializeHashContent(accB.Address)

	//just to bootstrap
	var shortHashA [8]byte
	var shortHashB [8]byte
	copy(shortHashA[:], hashA[0:8])
	copy(shortHashB[:], hashB[0:8])

	storage.State[shortHashA] = append(storage.State[shortHashA], &accA)
	storage.State[shortHashB] = append(storage.State[shortHashB], &accB)
}
