package bc

import (
	"golang.org/x/crypto/sha3"
	"time"
	"errors"
)

const (
	ProofSize = 9
)

type Block struct {
	Hash [32]byte
	PrevHash [32]byte
	Version uint8 //future updates
	Proof [ProofSize]byte //72-bit, enough even if the network gets really large
	Timestamp int64
	Difficulty uint8
	MerkleRoot [32]byte
	TxData []Transaction
}

//imitating constructor
func newBlock(prevBlock [32]byte) *Block {
	b := Block{}
	b.Version = 0x01
	b.PrevHash = prevBlock
	return &b
}

func (b *Block) addTx(tx Transaction) {

	//check if transaction is well-formed and enough funds are available
	if !(tx).VerifyTx() || uint64(tx.Info.Amount) > State[tx.Info.From].Balance {
		return
	}

	//indirection, because struct elements are "by value"
	accSender := State[tx.Info.From]
	accSender.TxCnt += 1
	accSender.Balance -= uint64(tx.Info.Amount)
	State[tx.Info.From] = accSender

	accReceiver := State[tx.Info.To]
	accReceiver.Balance += uint64(tx.Info.Amount)
	State[tx.Info.To] = accReceiver

	//b.TxData[serializeHashContent(tx.Info)] = *tx
	b.TxData = append(b.TxData, tx)
}

func (b *Block) finalizeBlock() {

	b.MerkleRoot = buildMerkleTree(b.TxData)
	b.Timestamp = time.Now().Unix()
	b.Difficulty = 20

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

func validateBlock(b *Block) bool {

	//prevHash check

	//check if enough bits are set to 0 as indicated in the "difficulty" field
	if !validateProofOfWork(b.Difficulty, b.Hash) {
		return false
	}

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
	if b.Hash != sha3.Sum256(append(proof,partialHashed[:]...)) {
		return false
	}

	//cmp merkle tree
	if buildMerkleTree(b.TxData) != b.MerkleRoot {
		return false
	}

	//check if transaction is syntactically well-formed and signature is correct
	for _, tx := range b.TxData {
		if !tx.VerifyTx() {
			return false
		}
	}

	//apply to State
	for _, tx := range b.TxData {
		if !stateChange(&tx) {
			stateRollBack()
		}
	}

	return true
}

func stateChange(tx *Transaction) error {
	//indirection, because struct elements are "by value"
	//does the address exist in the state?

	if _, exists := State[tx.Info.From]; !exists {
		return errors.New("Sender does not exist in the State.")
	}

	if _, exists := State[tx.Info.To]; !exists {
		return errors.New("Receiver does not exist in the State.")
	}

	if uint64(tx.Info.Amount) > State[tx.Info.From].Balance {
		return errors.New("Sender does not have enough funds for the transaction.")
	}

	accSender := State[tx.Info.From]
	accSender.TxCnt += 1
	accSender.Balance -= uint64(tx.Info.Amount)
	State[tx.Info.From] = accSender

	accReceiver := State[tx.Info.To]
	accReceiver.Balance += uint64(tx.Info.Amount)
	State[tx.Info.To] = accReceiver

	//all good
	return nil
}

func stateRollBack() {

}