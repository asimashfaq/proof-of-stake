package bc

import (
	"testing"
	"time"
	"math/rand"
	"encoding/binary"
)

func TestFundsStateChangeRollback(t *testing.T) {

	rand := rand.New(rand.NewSource(time.Now().Unix()))

	var testSize uint32
	testSize = 1000

	b := newBlock()
	var funds []*fundsTx

	var feeA, feeB uint64

	rollBackA := accA.Balance
	rollBackB := accB.Balance

	balanceA := accA.Balance
	balanceB := accB.Balance

	loopMax := int(rand.Uint32()%testSize+1)
	for i := 0; i < loopMax+1; i++ {
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
	getAccountFromHash(accA.Hash).TxCnt = 0
	getAccountFromHash(accB.Hash).TxCnt = 0
	fundsStateChange(funds)
	if accA.Balance != balanceA || accB.Balance != balanceB {
		t.Error("State update failed!")
	}
	fundsStateChangeRollback(funds)
	if accA.Balance != rollBackA || accB.Balance != rollBackB {
		t.Error("Rollback failed!")
	}
	minerBal := minerAcc.Balance
	collectTxFees(funds,nil,minerAcc.Hash)
	if feeA+feeB != minerAcc.Balance-minerBal {
		t.Error("Fee Collection failed!")
	}
	collectTxFeesRollback(funds,nil,minerAcc.Hash)
	if minerBal != minerAcc.Balance {
		t.Error("Fee Collection Rollback failed!")
	}
	balBeforeRew := minerAcc.Balance
	reward := 5
	collectBlockReward(uint64(reward),minerAcc.Hash)
	if minerAcc.Balance != balBeforeRew+uint64(reward) {
		t.Error("Block reward collection failed!")
	}
	collectBlockRewardRollback(uint64(reward),minerAcc.Hash)
	if minerAcc.Balance != balBeforeRew {
		t.Error("Block reward collection rollback failed!")
	}
}

func TestAccStateChangeRollback(t *testing.T) {

	rand := rand.New(rand.NewSource(time.Now().Unix()))

	var testSize uint32
	testSize = 1000

	var accs []*accTx

	loopMax := int(rand.Uint32()%testSize)+1
	for i := 0; i < loopMax; i++ {
		tx,_ := ConstrAccTx(rand.Uint64()%1000,&RootPrivKey)
		accs = append(accs, tx)
	}

	accStateChange(accs)

	var shortHash [8]byte
	for _,acc := range accs {
		found := false
		accHash := serializeHashContent(acc.PubKey)
		copy(shortHash[:],accHash[0:8])
		accSlice := State[shortHash]
		for _,singleAcc := range accSlice {
			if singleAcc.Hash == accHash {
				found = true
			}
		}
		if !found {
			t.Errorf("Account State failed to update for the following account: %v\n", acc)
		}
	}


	accStateChangeRollback(accs)

	for _,acc := range accs {
		found := false
		accHash := serializeHashContent(acc.PubKey)
		copy(shortHash[:],accHash[0:8])
		accSlice := State[shortHash]
		for _,singleAcc := range accSlice {
			if singleAcc.Hash == accHash {
				found = true
			}
		}
		if found {
			t.Errorf("Account State failed to rollback the following account: %v\n", acc)
		}
	}
}