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

var nextBlockAccess sync.Mutex

var txQueue, blockQueue *Queue

var timestamp []int64
var parameterSlice []parameters
var activeParameters *parameters

func Sync() {

}

func InitSystem() {

	txQueue = NewQueue()
	blockQueue = NewQueue()

	testing_setup()

	LogFile, _ = os.OpenFile("logminer "+time.Now().String(), os.O_RDWR|os.O_CREATE, 0666)
	log.SetOutput(LogFile)

	log.Println("Starting system, initializing state map")
	genesisBlock := newBlock()
	collectStatistics(genesisBlock)
	writeBlock(genesisBlock)

	currentBlock = newBlock()
	nextBlock = newBlock()

	go consumeBlock()
	go consumeTx()
	mining()
}

func consumeBlock() {

}



func publishBlock() {

}

func publishTx() {

}

func mining() {

	for {
		finalizeBlock(currentBlock)
		fmt.Print("Block mined.\n")
		if err := validateBlock(currentBlock); err != nil {
			fmt.Printf("%v\n", err)
		}

		nextBlockAccess.Lock()
		prevHash := currentBlock.Hash
		currentBlock = nextBlock
		currentBlock.PrevHash = prevHash
		//please no memory leaks :/
		nextBlock = newBlock()
		nextBlockAccess.Unlock()
	}
}

func InFundsTx(data []byte) {

}

func InAccTx(data []byte) {

	/*tx := DecodeAccTx(data)
	if tx == nil {
		return
	}
	txQueue.Enqueue(tx)*/
}

func InBlock(data []byte) {

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

	pub1, _ := new(big.Int).SetString(RootPub1, 16)
	pub2, _ := new(big.Int).SetString(RootPub2, 16)

	copy(pubKey[:32], pub1.Bytes())
	copy(pubKey[32:], pub2.Bytes())

	rootHash := serializeHashContent(pubKey)

	var shortRootHash [8]byte
	copy(shortRootHash[:], rootHash[0:8])
	rootAcc := protocol.Account{Address: pubKey}
	storage.State[shortRootHash] = append(storage.State[shortRootHash], &rootAcc)
	storage.RootKeys[rootHash] = &rootAcc
	var accA, accB protocol.Account

	puba1, _ := new(big.Int).SetString(pubA1, 16)
	puba2, _ := new(big.Int).SetString(pubA2, 16)
	priva, _ := new(big.Int).SetString(privA, 16)
	PubKeyA := ecdsa.PublicKey{
		elliptic.P256(),
		puba1,
		puba2,
	}
	PrivKeyA := ecdsa.PrivateKey{
		PubKeyA,
		priva,
	}

	pubb1, _ := new(big.Int).SetString(pubB1, 16)
	pubb2, _ := new(big.Int).SetString(pubB2, 16)
	privb, _ := new(big.Int).SetString(privB, 16)
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
