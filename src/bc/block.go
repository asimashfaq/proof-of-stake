package bc

import (
	"golang.org/x/crypto/sha3"
	"time"
	"errors"
	"log"
	"fmt"
	"bytes"
	"encoding/binary"
)

const (
	ProofSize = 9
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
func newBlock(prevBlockHash [32]byte) *Block {
	b := Block{}
	b.Version = 0x01
	b.PrevHash = prevBlockHash
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
		log.Printf("Transaction could not be verified: %v\n", tx)
		return errors.New("Transaction could not be verified.")
	}

	switch tx.(type) {
	case *fundsTx:
		err := b.addFundsTx(tx.(*fundsTx))
		if err != nil {
			log.Printf("Adding fundsTx tx failed (%v): %v\n",err, tx.(*fundsTx))
		}
	case *accTx:
		err := b.addAccTx(tx.(*accTx))
		if err != nil {
			log.Printf("Adding accTx tx failed: %x, because %v\n", tx.(*accTx),err)
		}
	default:
		return errors.New("Transaction type not recognized.")
	}

	return nil
}

func (b *Block) addAccTx(tx *accTx) error {

	//at this point the tx has already been verified
	var mapId [8]byte
	accHash := sha3.Sum256(tx.PubKey[:])
	copy(mapId[:],accHash[0:8])
	for _,j := range State[mapId] {
		if bytes.Compare(tx.PubKey[:],j.Address[:]) == 0 {
			return errors.New("Account already exists.")
		}
	}

	b.AccTxData = append(b.AccTxData,*tx)
	log.Printf("Added tx to the AccTxData slice: %v", *tx)
	return nil
}

func (b *Block) addFundsTx(tx *fundsTx) error {

	//checking if the sender account is already in the local state copy
	if _,exists := b.stateCopy[tx.fromHash]; !exists {
		for _,acc := range State[tx.From] {
			if bytes.Compare(acc.Hash[:],tx.fromHash[:]) == 0 {
				b.stateCopy[tx.fromHash] = acc
			}
		}
	}

	//vice versa for receiver account
	if _,exists := b.stateCopy[tx.toHash]; !exists {
		for _,acc := range State[tx.To] {
			if bytes.Compare(acc.Hash[:],tx.toHash[:]) == 0 {
				b.stateCopy[tx.toHash] = acc
			}
		}
	}

	amount := binary.BigEndian.Uint32(tx.Amount[:])
	if uint64(amount) > b.stateCopy[tx.fromHash].Balance {
		return errors.New("Not enough funds to complete the transaction")
	}

	accSender := b.stateCopy[tx.fromHash]
	accSender.TxCnt += 1
	accSender.Balance -= uint64(amount)
	b.stateCopy[tx.fromHash] = accSender

	accReceiver := b.stateCopy[tx.toHash]
	accReceiver.Balance += uint64(amount)
	b.stateCopy[tx.toHash] = accReceiver

	//b.TxData[serializeHashContent(tx.Info)] = *tx
	b.FundsTxData = append(b.FundsTxData, *tx)

	log.Printf("Added tx to the block FundsTxData slice: %v", *tx)
	return nil
}

func (b *Block) finalizeBlock() {

	//merkle tree only built from funds transactions
	b.MerkleRoot = buildMerkleTree(b.FundsTxData)
	b.Timestamp = time.Now().Unix()
	b.Difficulty = 18

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
	log.Printf("Finalized block: %v", b)
}

func validateBlock(b *Block) error {

	//prevHash check
	//extract proof first by cutting of leading zeroes
	log.Println("Starting block validation...")
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

	log.Println("Proof of work validation passed.")

	//cmp merkle tree
	if buildMerkleTree(b.FundsTxData) != b.MerkleRoot {
		return errors.New("Merkle Root incorrect.")
	}

	log.Println("Merkle root hash passed.")


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
	for index, tx := range b.FundsTxData {
		if fundsStateChange(&tx) != nil {
			//don't use pointer here
			log.Println("Starting rollback")
			fundsStateRollback(b.FundsTxData,index-1)
			return errors.New("Invalid State Transition. Roll back.")
		}
	}

	for _,tx := range b.AccTxData {
		accStateChange(&tx)
	}

	return nil
}

func (b Block) String() string {
	return fmt.Sprintf("\nHash: %x\n" +
		"Previous Hash: %x\n" +
		"Version: %v\n" +
		"Proof: %x\n" +
		"Timestamp: %v\n" +
		"Difficulty: %v\n" +
		"MerkleRoot: %x\n" +
		"Amount of fundsTx: %v\n" +
		"Amount of txData: %v\n",
		b.Hash[0:4],
		b.PrevHash[0:4],
		b.Version,
		b.Proof,
		b.Timestamp,
		b.Difficulty,
		b.MerkleRoot[0:4],
		len(b.FundsTxData),
		len(b.AccTxData),
	)
}