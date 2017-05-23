package bc

import (
	"crypto/ecdsa"
	"log"
	"os"
	"time"
	"math/big"
	"crypto/elliptic"
	"crypto/rand"
)

//will act as interface to bc package
var State map[[8]byte][]*Account
var RootKeys map[[32]byte]*Account
var LogFile *os.File
var block *Block

var MinerHash [32]byte
var MinerPrivKey *ecdsa.PrivateKey

func Sync(){

}

func InitSystem() {

	State = make(map[[8]byte][]*Account)
	RootKeys = make(map[[32]byte]*Account)

	LogFile, _ = os.OpenFile("log "+time.Now().String(), os.O_RDWR | os.O_CREATE , 0666)
	log.SetOutput(LogFile)

	//set up mining account
	MinerPrivKey, _ = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	var pubKey [64]byte
	var shortMiner [8]byte
	copy(pubKey[:32],MinerPrivKey.X.Bytes())
	copy(pubKey[32:],MinerPrivKey.Y.Bytes())
	MinerHash = serializeHashContent(pubKey[:])
	copy(shortMiner[:],MinerHash[0:8])
	minerAcc := Account{Hash:MinerHash, Address:pubKey}
	State[shortMiner] = append(State[shortMiner],&minerAcc)

	log.Println("Starting system, initializing state map")
	//temporary
	block = newBlock([32]byte{})

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
}

func AddFundsTx(localTxCnt uint32, from, to [32]byte, amount uint32, key *ecdsa.PrivateKey) (error) {
	/*var header byte
	var buf bytes.Buffer
	var amountBuf [4]byte
	var txCntBuf [3]byte
	var feeBuf [2]byte
	var fee uint16
	fee = 1

	//this has to be easier
	var tmpTxCntBuf [4]byte

	binary.Write(&buf, binary.BigEndian, localTxCnt)
	copy(tmpTxCntBuf[:],buf.Bytes())
	copy(txCntBuf[:],tmpTxCntBuf[1:])
	buf.Reset()

	binary.Write(&buf, binary.BigEndian, fee)
	copy(feeBuf[:],buf.Bytes())
	buf.Reset()
/*var header byte
	var buf bytes.Buffer
	var amountBuf [4]byte
	var txCntBuf [3]byte
	var feeBuf [2]byte
	var fee uint16
	fee = 1

	//this has to be easier
	var tmpTxCntBuf [4]byte

	binary.Write(&buf, binary.BigEndian, localTxCnt)
	copy(tmpTxCntBuf[:],buf.Bytes())
	copy(txCntBuf[:],tmpTxCntBuf[1:])
	buf.Reset()

	binary.Write(&buf, binary.BigEndian, fee)
	copy(feeBuf[:],buf.Bytes())
	buf.Reset()

	binary.Write(&buf, binary.BigEndian, amount)
	copy(amountBuf[:],buf.Bytes())
	buf.Reset()

	tx, err := constrFundsTx(header, amountBuf, feeBuf, txCntBuf, from,to, key)
	fmt.Printf("%v\n", tx)
	//serialize tx

	data := encodeFundsTx(tx)
	binary.Write(&buf, binary.BigEndian, amount)
	copy(amountBuf[:],buf.Bytes())
	buf.Reset()

	tx, err := constrFundsTx(header, amountBuf, feeBuf, txCntBuf, from,to, key)
	fmt.Printf("%v\n", tx)
	//serialize tx

	data := encodeFundsTx(tx)


	decodeData(data)

	//localTxCnt++
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	err = block.addTx(&tx)
	if err != nil {
		fmt.Printf("%v\n", err)
	}*/
	return nil
}

func AddAccTx() *accTx {

	tx,err := constrAccTx()

	if err != nil {
		log.Printf("%v\n", err)
		//return errors.New("Failed to construct account tx.")
	}
	block.addTx(&tx)
	//return nil
	return &tx
}

//temporary
func FinalizeBlock() {
	block.finalizeBlock()
}

func ValidateBlock() {

	if validateBlock(block) != nil {
		return
	}
	prevBlock := block
	block = newBlock(prevBlock.Hash)
}

//gets called from the main network receiver loop
func decodeData(payload []byte) {

	switch(len(payload)) {
	//fixed length input packets
	case 90:
		//_fundsTx := decodeFundsTx(payload)

	}


}