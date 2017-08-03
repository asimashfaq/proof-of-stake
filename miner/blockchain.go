package miner

import (
	"fmt"
	"github.com/lisgie/bazo_miner/protocol"
	"github.com/lisgie/bazo_miner/storage"
	"log"
	"math/big"
	"os"
	"sync"
	"time"
)

var (
	logger               *log.Logger
	blockValidation      = &sync.Mutex{}
	parameterSlice       []parameters
	activeParameters     *parameters
	uptodate             bool
)

//Miner entry point
func Init() {

	//Initialize root key
	initRootKey()

	//Set up logger
	LogFile, _ := os.OpenFile("log/miner "+time.Now().String(), os.O_RDWR|os.O_CREATE, 0666)
	logger = log.New(LogFile, "", log.LstdFlags)

	parameterSlice = append(parameterSlice, parameters{
		[32]byte{},
		1,
		5000000, //5MB
		2016,
		60, //1min
		0,
	})
	activeParameters = &parameterSlice[0]

	currentTargetTime = new(timerange)
	target = append(target, 13)

	//Start blockchain with genesis block and 0 hash
	//Don't validate nor broadcast
	genesis := newBlock([32]byte{})
	collectStatistics(genesis)
	storage.WriteClosedBlock(genesis)

	//Start to listen to network inputs (txs and blocks)
	go incomingData()
	mining()
}

//Mining is a constant process, trying to come up with a successful PoW
func mining() {
	currentBlock := newBlock([32]byte{})
	for {
		err := finalizeBlock(currentBlock)
		if err != nil {
			fmt.Printf("%v\n", err)
		} else {
			fmt.Println("Block mined.")
		}
		//else a block was received meanwhile that was added to the chain, all the effort was in vain :(
		//wait for lock here only
		if err != nil {
			logger.Printf("%v\n", err)
		} else {
			broadcastBlock(currentBlock)
			err := validateBlock(currentBlock)
			if err != nil {
				logger.Printf("Received block (%x) could not be validated: %v\n", currentBlock.Hash[0:12], err)
			}
		}

		//This is the same mutex that is claimed at the beginning of a block validation. The reason we do this is
		//that before start mining a new block we empty the mempool which contains tx data that is likely to be
		//validated with block validation, so we wait in order to not work on tx data that is already validated
		//when we finish the block
		blockValidation.Lock()
		nextBlock := newBlock(lastBlock.Hash)
		currentBlock = nextBlock
		prepareBlock(currentBlock)
		blockValidation.Unlock()
	}
}

//At least one root key needs to be set which is allowed to create new accounts
func initRootKey() {

	var pubKey [64]byte

	pub1, _ := new(big.Int).SetString(INITROOTKEY1, 16)
	pub2, _ := new(big.Int).SetString(INITROOTKEY2, 16)

	copy(pubKey[:32], pub1.Bytes())
	copy(pubKey[32:], pub2.Bytes())

	rootHash := serializeHashContent(pubKey)

	rootAcc := protocol.Account{Address: pubKey}

	storage.State[rootHash] = &rootAcc
	storage.RootKeys[rootHash] = &rootAcc
}
