package bc

import (
	"crypto/ecdsa"
	"log"
	"os"
	"time"
	"math/big"
	"crypto/elliptic"
	"crypto/rand"
	"golang.org/x/crypto/sha3"
	"sync"
	"fmt"
)

const (
	pubA1 = "c2be9abbeaec39a066c2a09cee23bb9ab2a0b88f2880b1e785b4d317adf0dc7c"
	pubA2 = "8ce020fde838d9c443f6c93345dafe7fd74f091c4d2f30b37e2453679a257ed5"
	privA = "ba127fa8f802b008b9cdb58f4e44809d48f1b000cff750dda9cd6b312395c1c5"
	pubB1 = "5d7eefd58e3d2f309471928ab4bbd104e52973372c159fa652b8ca6b57ff68b8"
	pubB2 = "ab301a6a77b201c416ddc13a2d33fdf200a5302f6f687e0ea09085debaf8a1d9"
	privB = "7a0a9babcc97ea7991ed67ed7f800f70c5e04e99718960ad8efab2ca052f00c7"
)

//will act as interface to bc package
var State map[[8]byte][]*Account
var RootKeys map[[32]byte]*Account
var LogFile *os.File
var currentBlock, nextBlock *Block

var MinerHash [32]byte
var MinerPrivKey *ecdsa.PrivateKey

var nextBlockAccess sync.Mutex

var txQueue,blockQueue *Queue

func Sync(){

}

func InitSystem() {
	txQueue = NewQueue()
	blockQueue = NewQueue()

	State = make(map[[8]byte][]*Account)
	RootKeys = make(map[[32]byte]*Account)

	testing_setup()

	LogFile, _ = os.OpenFile("log "+time.Now().String(), os.O_RDWR | os.O_CREATE , 0666)
	log.SetOutput(LogFile)

	//set up mining account

	log.Println("Starting system, initializing state map")
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
			nextBlock.addTx(txQueue.Dequeue().(transaction))
			nextBlockAccess.Unlock()
		}
		time.Sleep(200*time.Millisecond)
	}
}

func publishBlock() {

}

func publishTx() {

}

func mining() {

	for {
		currentBlock.finalizeBlock()
		fmt.Print("Block mined.\n")
		validateBlock(currentBlock)

		nextBlockAccess.Lock()
		prevHash := currentBlock.Hash
		currentBlock = nextBlock
		currentBlock.PrevHash = prevHash
		//please no memory leaks :/
		nextBlock = newBlock()
		nextBlockAccess.Unlock()
	}
}

func ProcessInput(data []byte) {

	//inspect header
	//parse input (what kind of tx, block etc.)
	tx := DecodeFundsTx(data[1:])
	processFundsTx(tx)
}

func processFundsTx(tx* fundsTx) {

	txQueue.Enqueue(tx)
	//nextBlock.addTx(tx)
}

func AddFundsTx(localTxCnt uint32, from, to [32]byte, amount uint32, key *ecdsa.PrivateKey) (error) {
	return nil
}

func ValidateBlock() {

}

//gets called from the main network receiver loop
func decodeData(payload []byte) {

	switch(len(payload)) {
	//fixed length input packets
	case 90:
		//_fundsTx := decodeFundsTx(payload)

	}
}

//some testing code
func testing_setup() {
	MinerPrivKey, _ = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	var pubKey [64]byte
	var shortMiner [8]byte
	copy(pubKey[:32],MinerPrivKey.X.Bytes())
	copy(pubKey[32:],MinerPrivKey.Y.Bytes())
	MinerHash = serializeHashContent(pubKey[:])
	copy(shortMiner[:],MinerHash[0:8])
	minerAcc := Account{Hash:MinerHash, Address:pubKey}
	State[shortMiner] = append(State[shortMiner],&minerAcc)

	pub1,_ := new(big.Int).SetString(RootPub1,16)
	pub2,_ := new(big.Int).SetString(RootPub2,16)

	copy(pubKey[:32],pub1.Bytes())
	copy(pubKey[32:],pub2.Bytes())

	rootHash := serializeHashContent(pubKey[:])

	var shortRootHash [8]byte
	copy(shortRootHash[:], rootHash[0:8])
	rootAcc := Account{Hash:rootHash, Address:pubKey}
	State[shortRootHash] = append(State[shortRootHash], &rootAcc)
	RootKeys[rootHash] = &rootAcc
	var accA, accB Account

	puba1,_ := new(big.Int).SetString(pubA1,16)
	puba2,_ := new(big.Int).SetString(pubA2,16)
	priva,_ := new(big.Int).SetString(privA,16)
	PubKeyA := ecdsa.PublicKey{
		elliptic.P256(),
		puba1,
		puba2,
	}
	PrivKeyA := ecdsa.PrivateKey{
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

	accA = Account{Balance: 15000}
	copy(accA.Address[0:32], PrivKeyA.PublicKey.X.Bytes())
	copy(accA.Address[32:64], PrivKeyA.PublicKey.Y.Bytes())
	accA.Hash = sha3.Sum256(accA.Address[:])

	//This one is just for testing purposes
	accB = Account{Balance: 702}
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
