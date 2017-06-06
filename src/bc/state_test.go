package bc

import (
	"math/rand"
	"testing"
	"encoding/binary"
	"time"
)

//Testing state change, rollback and fee collection
func TestFundsTxStateChange(t *testing.T) {

	rand := rand.New(rand.NewSource(time.Now().Unix()))

	var testSize uint32
	testSize = 1000000

	b := newBlock()
	var funds []*fundsTx

	var feeA, feeB uint64

	rollBackA := accA.Balance
	rollBackB := accB.Balance

	balanceA := accA.Balance
	balanceB := accB.Balance

	for i := 0; i < int(rand.Uint32()%testSize+1); i++ {
		ftx, _ := ConstrFundsTx(0x01,rand.Uint64()%1000000+1, rand.Uint64()%100+1, uint32(i), accA.Hash, accB.Hash, &PrivKeyA)
		if b.addTx(ftx) == nil {
			funds = append(funds,ftx)
			amount := binary.BigEndian.Uint64(ftx.Amount[:])
			fee := binary.BigEndian.Uint64(ftx.Fee[:])
			balanceA -= amount
			feeA += fee

			balanceB += amount
		}



		ftx2,_ := ConstrFundsTx(0x01,rand.Uint64()%1000+1, rand.Uint64()%100+1, uint32(i), accB.Hash, accA.Hash, &PrivKeyB)
		if b.addTx(ftx2) == nil {
			funds = append(funds,ftx2)
			amount := binary.BigEndian.Uint64(ftx2.Amount[:])
			fee := binary.BigEndian.Uint64(ftx2.Fee[:])
			balanceB -= amount
			feeB += fee

			balanceA += amount
		}
	}

	for _,tx := range funds {
		fundsStateChange(tx)
	}

	if accA.Balance != balanceA || accB.Balance != balanceB {
		t.Error("State update failed!")
	}

	fundsStateRollback(funds, len(funds)-1)

	if accA.Balance != rollBackA || accB.Balance != rollBackB {
		t.Error("Rollback failed!")
	}

	minerBal := minerAcc.Balance
	collectTxFees(funds,nil,minerAcc.Hash)
	if feeA+feeB != minerAcc.Balance-minerBal {
		t.Error("Fee Collection failed!")
	}
}

func TestAccTxStateChange(t *testing.T) {

	rand := rand.New(rand.NewSource(time.Now().Unix()))

	var testSize uint32
	testSize = 10

	var accs []*accTx


	for i := 0; i < int(rand.Uint32()%testSize)+1; i++ {
		tx,_ := ConstrAccTx(rand.Uint64()%1000,&RootPrivKey)
		accs = append(accs, tx)
	}

	for _,tx := range accs {
		accStateChange(tx)
	}

	var shortHash [8]byte
	for _,acc := range accs {
		accHash := serializeHashContent(acc.PubKey)
		copy(shortHash[:],accHash[0:8])
		if _,exists := State[shortHash]; !exists {
			t.Errorf("Account State failed to update for the following account: %v\n", acc)
		}
	}
}