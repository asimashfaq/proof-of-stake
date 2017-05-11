package bc

import (
	"golang.org/x/crypto/sha3"
	"time"
	"errors"
)

const (
	ProofSize = 19
)

type transaction interface {
	verify() bool
}

type Block struct {
	Hash [32]byte
	PrevHash [32]byte
	Version uint8 //future updates
	//72-bit, enough even if the network gets really large
	Proof [ProofSize]byte
	Timestamp int64
	Difficulty uint8
	MerkleRoot [32]byte
	//this field will not be exported, this is just to avoid race conditions for the global state
	stateCopy map[[32]byte]Account
	FundsTxData []fundsTx
	AccTxData []accTx
}

//imitating constructor
func newBlock(prevBlock [32]byte) *Block {
	b := Block{}
	b.Version = 0x01
	b.PrevHash = prevBlock
	b.stateCopy = make(map[[32]byte]Account)
	return &b
}

//this method is to validate transactions, a copy of the state
// is used for every instead of manipulating the global state
//because we the work might get interrupted by receiving a block
func (b *Block) addTx(tx transaction) error {

	//verifies correctness for the specific transaction
	//i'd actually like to use !(&tx).verify to pass by pointer, but golang doesn't allow this
	if !(tx).verify() {
		return errors.New("Transaction could not be verified.")
	}

	switch tx.(type) {
	case *fundsTx:
		b.addFundsTx(tx)
	case *accTx:
		b.addAccTx(tx)
	default:
		return errors.New("Transaction type not recognized.")
	}

	return nil
}

func (b *Block) addAccTx(tx transaction) error {

	//accTx := tx.(*accTx)

	//at this point the tx has already been verified
	//if _,exists := State[]

	return nil
}

func (b *Block) addFundsTx(tx transaction) error {

	fundsTx := tx.(*fundsTx)

	if _,exists := b.stateCopy[fundsTx.Payload.To]; !exists {
		b.stateCopy[fundsTx.Payload.To] = State[fundsTx.Payload.To]
	}

	if _,exists := b.stateCopy[fundsTx.Payload.From]; !exists {
		b.stateCopy[fundsTx.Payload.From] = State[fundsTx.Payload.From]
	}

	if uint64(fundsTx.Payload.Amount) > State[fundsTx.Payload.From].Balance {
		return errors.New("Not enough funds to complete the transaction")
	}

	accSender := b.stateCopy[fundsTx.Payload.From]
	accSender.TxCnt += 1
	accSender.Balance -= uint64(fundsTx.Payload.Amount)
	b.stateCopy[fundsTx.Payload.From] = accSender

	b.stateCopy[fundsTx.Payload.To] = State[fundsTx.Payload.To]
	accReceiver := b.stateCopy[fundsTx.Payload.To]
	accReceiver.Balance += uint64(fundsTx.Payload.Amount)
	b.stateCopy[fundsTx.Payload.To] = accReceiver

	//b.TxData[serializeHashContent(tx.Info)] = *tx
	b.FundsTxData = append(b.FundsTxData, *fundsTx)
	return nil

}

func (b *Block) finalizeBlock() {

	//merkle tree only built from funds transactions
	b.MerkleRoot = buildMerkleTree(b.FundsTxData)
	b.Timestamp = time.Now().Unix()
	b.Difficulty = 8

	//anonymous struct
	partialToHash := struct{
		prevHash [32]byte
		version uint8
		timestamp int64
		merkleRoot [32]byte
	}{
		b.PrevHash,
		b.Version,
		b.Timestamp,
		b.MerkleRoot,
	}

	partialHashed := serializeHashContent(partialToHash)
	proof := proofOfWork(b.Difficulty, partialHashed)
	b.Hash = sha3.Sum256(append(proof.Bytes(),partialHashed[:]...))

	//we need to write the proof at the end of the fixed-size byte array of length 9
	//needs to be decoded by the receiver
	for index,val := range proof.Bytes() {
		b.Proof[ProofSize-len(proof.Bytes())+index] = val
	}
}

func validateBlock(b *Block) error {

	//prevHash check
	//extract proof first by cutting of leading zeroes
	startIndex := 0
	for _, singleByte := range b.Proof {
		if singleByte != 0x00 {
			break
		}
		startIndex++
	}
	proof := b.Proof[startIndex:]

	//anonymous struct17
	partialToHash := struct{
		prevHash [32]byte
		version uint8
		timestamp int64
		merkleRoot [32]byte
	}{
		b.PrevHash,
		b.Version,
		b.Timestamp,
		b.MerkleRoot,
	}
	partialHashed := serializeHashContent(partialToHash)
	if b.Hash != sha3.Sum256(append(proof,partialHashed[:]...)) || !validateProofOfWork(b.Difficulty, b.Hash) {
		return errors.New("Proof of work is incorrect.")
	}

	//cmp merkle tree
	if buildMerkleTree(b.FundsTxData) != b.MerkleRoot {
		return errors.New("Merkle Root incorrect.")
	}

	//check if fundsTxs is syntactically well-formed and signature is correct
	for _, tx := range b.FundsTxData {
		if !tx.verify() {
			return errors.New("Malformed transaction.")
		}
	}

	//check if accTxs are syntactically well-formed and signature is correct
	for _, tx := range b.AccTxData {
		if !tx.verify() {
			return errors.New("Malformed transaction.")
		}
	}

	//apply to State
	/*for index, tx := range b.FundsTxData {
		if stateChange(&tx) != nil {
			stateRollBack(index-1, b.FundsTxData)
			return errors.New("Invalid State Transition. Roll back.")
		}
	}*/

	return nil
}