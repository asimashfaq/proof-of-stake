package bc

import (
	"math/big"
	"golang.org/x/crypto/sha3"
	"time"
)

type Block struct {
	Hash [32]byte
	PrevHash [32]byte
	Version uint8 //future updates
	Timestamp int64
	MerkleRoot [32]byte
	TxData map[[32]byte]Transaction
	StateCopy map[[64]byte]Account
}

//imitating constructor
func NewBlock(prevBlock [32]byte, stateCopy map[[64]byte]Account) *Block {
	b := Block{StateCopy:stateCopy}
	b.TxData = make(map[[32]byte]Transaction)
	b.Version = 0x01
	b.PrevHash = prevBlock
	return &b
}

func (b *Block) AddTx(tx *Transaction) {

	//check if transaction is well-formed and enough funds are available
	if !(*tx).VerifyTx() || tx.Info.Amount > b.StateCopy[tx.Info.From].Balance {
		return
	}

	//state change, only the sender's txcnt is increased
	b.StateCopy[tx.Info.From] = Account{
		b.StateCopy[tx.Info.From].TxCnt + 1,
		b.StateCopy[tx.Info.From].Balance - tx.Info.Amount,
	}

	b.StateCopy[tx.Info.To] = Account{
		b.StateCopy[tx.Info.To].TxCnt,
		b.StateCopy[tx.Info.To].Balance + tx.Info.Amount,
	}
	b.TxData[serializeHashContent(tx.Info)] = *tx
}

func (b *Block) FinalizeBlock() {

	b.MerkleRoot = buildMerkleTree(b.TxData)
	//proof := proofOfWork(24, b.MerkleRoot)
	b.Timestamp = time.Now().Unix()

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