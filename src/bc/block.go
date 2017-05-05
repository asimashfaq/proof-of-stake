package bc

import (
	"fmt"
	"math/big"
	"golang.org/x/crypto/sha3"
)

type Block struct {
	Hash [32]byte
	PrevHash [32]byte
	Version uint8
	Timestamp int64
	NrOfTransactions int32
	TxData map[[32]byte]Transaction //slice
	StateCopy map[[64]byte]int64
}

//imitating constructor
func NewBlock(stateCopy map[[64]byte]int64) *Block {
	b := Block{StateCopy:stateCopy}
	b.TxData = make(map[[32]byte]Transaction)
	return &b
}

func (b *Block) AddTx(tx *Transaction) {

	//check if transaction is well-formed and enough funds are available
	if !(*tx).VerifyTx() || tx.Info.Amount > b.StateCopy[tx.Info.From] {
		return
	}

	//state change
	b.StateCopy[tx.Info.From] -= tx.Info.Amount
	b.StateCopy[tx.Info.To] += tx.Info.Amount
	b.NrOfTransactions++

	b.TxData[sha3.Sum256(serializeTxContent(tx.Info))] = *tx
}

func (b *Block) FinalizeBlock() {


	merkleRoot := buildMerkleTree(b.TxData)
	fmt.Printf("%x\n", sha3.Sum256(append(proofOfWork(24, merkleRoot).Bytes(),merkleRoot[:]...)))

}

func proofOfWork(diff uint8, merkleRoot [32]byte) *big.Int {

	var tmp [32]byte
	var byteNr uint8
	var abort bool
	//big int needed because int64 overflows if nonce too large
	oneIncr := big.NewInt(1)
	cnt := big.NewInt(0)

	for ;; cnt.Add(cnt,oneIncr) {
		abort = false

		tmp = sha3.Sum256(append(cnt.Bytes(),merkleRoot[:]...))
		for byteNr = 0; byteNr < (uint8)(diff/8); byteNr++ {
			if tmp[byteNr] != 0 {
				abort = true
				break
			}
		}
		if abort {
			continue
		}

		if diff%8 != 0 && tmp[byteNr+1] >= 1<<(8-diff%8) {
			continue
		}
		break
	}

	return cnt
}