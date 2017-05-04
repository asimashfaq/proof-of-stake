package bc

import (
	"fmt"
	"math/big"
	"golang.org/x/crypto/sha3"
	"strconv"
)

type Block struct {
	Hash [32]byte
	PrevHash [32]byte
	Version uint8
	Timestamp int64
	NrOfTransactions int32
	Data []Transaction //slice
	StateCopy map[[64]byte]int64
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
	b.Data = append(b.Data, *tx)
}

func (b *Block) FinalizeBlock() {

	fmt.Print("")
	/*for _, j := range b.Data {
		fmt.Printf("%x\n", j)
	}

	for j := range b.StateCopy {
		fmt.Printf("%x: %d\n", j, b.StateCopy[j])
	}*/
}

func proofOfWork(diff uint8, merkleRoot [32]byte) *big.Int {

	var tmp [32]byte
	var bits uint8
	var byteNr uint8
	var abort bool

	for i := 0 ;; i++ {
		abort = false
		tmp = sha3.Sum256(append(merkleRoot[:],[]byte(strconv.Itoa(i))...))

		fmt.Printf("%x\n", tmp)

		for byteNr = 0; byteNr < (uint8)(bits/8); byteNr++ {
			if tmp[byteNr] != 0 {
				abort = true
				break
			}
		}
		if abort {
			continue
		}

		if bits%8 != 0 && tmp[byteNr+1] >= 1<<(8-bits%8) {
			continue
		}

		fmt.Printf("%x\n", tmp)

		break
	}


	return nil
}