package bc

type Block struct {
	Hash [32]byte
	PrevHash [32]byte
	Version uint8
	Timestamp int64
	NrOfTransactions int32
	data []Transaction
	StateCopy map[[64]byte]int64
}

func (b *Block) AddTx(tx *Transaction) {
	if !tx.VerifyTx() || tx.Info.Amount > b.StateCopy[tx.Info.From] {
		return
	}

	//state change
	b.StateCopy[tx.Info.From] -= tx.Info.Amount
	b.StateCopy[tx.Info.To] += tx.Info.Amount
	b.NrOfTransactions++
}

func (b *Block) FinalizeBlock() {

}