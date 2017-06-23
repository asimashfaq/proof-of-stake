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

const (
	pubA1 = "c2be9abbeaec39a066c2a09cee23bb9ab2a0b88f2880b1e785b4d317adf0dc7c"
	pubA2 = "8ce020fde838d9c443f6c93345dafe7fd74f091c4d2f30b37e2453679a257ed5"
	privA = "ba127fa8f802b008b9cdb58f4e44809d48f1b000cff750dda9cd6b312395c1c5"
	pubB1 = "5d7eefd58e3d2f309471928ab4bbd104e52973372c159fa652b8ca6b57ff68b8"
	pubB2 = "ab301a6a77b201c416ddc13a2d33fdf200a5302f6f687e0ea09085debaf8a1d9"
	privB = "7a0a9babcc97ea7991ed67ed7f800f70c5e04e99718960ad8efab2ca052f00c7"
)

const (
	//P-256
	RootPub1 = "6323cc034597195ae69bcfb628ecdffa5989c7503154c566bab4a87f3e9910ac"
	RootPub2 = "f6115b77a15852764c609c6a5c1739e698ebc6e49bf14617c561b9110039cec7"
	RootPriv = "277ed539f56122c25a6fc115d07d632b47e71416c9aebf1beb54ee704f11842c"
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

func consumeTx() {

	for {
		if txQueue.Size() != 0 {
			nextBlockAccess.Lock()
			addTx(nextBlock, txQueue.Dequeue().(protocol.Transaction))
			nextBlockAccess.Unlock()
		}
		time.Sleep(20 * time.Millisecond)
	}
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
