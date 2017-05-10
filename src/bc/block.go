package bc

import (
	"golang.org/x/crypto/sha3"
	"time"
	"errors"
)

const (
	ProofSize = 19
)

type transaction bool {}

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
	txData []transaction
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
func (b *Block) addTx(tx Transaction) error {

	if _,exists := b.stateCopy[tx.Payload.To]; !exists {
		b.stateCopy[tx.Payload.To] = State[tx.Payload.To]
	}

	if _,exists := b.stateCopy[tx.Payload.From]; !exists {
		b.stateCopy[tx.Payload.From] = State[tx.Payload.From]
	}

	//check if transaction is well-formed and enough funds are available
	if !(tx).VerifyTx() {
		return errors.New("fundsTx could not be verified.")
	}

	if uint64(tx.Payload.Amount) > State[tx.Payload.From].Balance {
		return errors.New("Not enough funds to complete the transaction")
	}

	accSender := b.stateCopy[tx.Payload.From]
	accSender.TxCnt += 1
	accSender.Balance -= uint64(tx.Payload.Amount)
	b.stateCopy[tx.Payload.From] = accSender

	b.stateCopy[tx.Payload.To] = State[tx.Payload.To]
	accReceiver := b.stateCopy[tx.Payload.To]
	accReceiver.Balance += uint64(tx.Payload.Amount)
	b.stateCopy[tx.Payload.To] = accReceiver

	//b.TxData[serializeHashContent(tx.Info)] = *tx
	b.TxData = append(b.TxData, tx)
	return nil
}

func (b *Block) finalizeBlock() {

	b.MerkleRoot = buildMerkleTree(b.TxData)
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
	if buildMerkleTree(b.TxData) != b.MerkleRoot {
		return errors.New("Merkle Root incorrect.")
	}

	//check if transaction is syntactically well-formed and signature is correct
	for _, tx := range b.TxData {
		if !tx.VerifyTx() {
			return errors.New("Malformed transaction.")
		}
	}

	//apply to State
	for index, tx := range b.TxData {
		if stateChange(&tx) != nil {
			stateRollBack(index-1, b.TxData)
			return errors.New("Invalid State Transition. Roll back.")
		}
	}

	return nil
}

func stateChange(tx *Transaction) error {

	if _, exists := State[tx.Payload.From]; !exists {
		return errors.New("Sender does not exist in the State.")
	}

	if _, exists := State[tx.Payload.To]; !exists {
		return errors.New("Receiver does not exist in the State.")
	}

	if uint64(tx.Payload.Amount) > State[tx.Payload.From].Balance {
		return errors.New("Sender does not have enough funds for the transaction.")
	}

	accSender := State[tx.Payload.From]
	accSender.TxCnt += 1
	accSender.Balance -= uint64(tx.Payload.Amount)
	State[tx.Payload.From] = accSender

	accReceiver := State[tx.Payload.To]
	accReceiver.Balance += uint64(tx.Payload.Amount)
	State[tx.Payload.To] = accReceiver

	//all good
	return nil
}

func stateRollBack(index int, txData []Transaction) {

	//in case the first entry failed we don't need to rollback
	if index == -1 {
		return
	}

	for cnt := index; cnt >= 0; cnt-- {
		tx := txData[cnt]
		accSender := State[tx.Payload.From]
		accSender.TxCnt -= 1
		accSender.Balance += uint64(tx.Payload.Amount)
		State[tx.Payload.From] = accSender

		accReceiver := State[tx.Payload.To]
		accReceiver.Balance -= uint64(tx.Payload.Amount)
		State[tx.Payload.To] = accReceiver
	}
}