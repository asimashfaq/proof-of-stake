package bc

import (
	"testing"
	"time"
	"math/rand"
)

func TestFundsStateChangeRollback(t *testing.T) {

	rand := rand.New(rand.NewSource(time.Now().Unix()))

	accAHash := serializeHashContent(accA.Address)
	accBHash := serializeHashContent(accB.Address)
	minerAccHash := serializeHashContent(minerAcc.Address)

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
		ftx, _ := ConstrFundsTx(0x01,rand.Uint64()%1000000+1, rand.Uint64()%100+1, uint32(i), accAHash, accBHash, &PrivKeyA)
		if b.addTx(ftx) == nil {
			funds = append(funds,ftx)
			balanceA -= ftx.Amount
			feeA += ftx.Fee

			balanceB += ftx.Amount
		}

		ftx2,_ := ConstrFundsTx(0x01,rand.Uint64()%1000+1, rand.Uint64()%100+1, uint32(i), accBHash, accAHash, &PrivKeyB)
		if b.addTx(ftx2) == nil {
			funds = append(funds,ftx2)
			balanceB -= ftx2.Amount
			feeB += ftx2.Fee

			balanceA += ftx2.Amount
		}
	}
	getAccountFromHash(accAHash).TxCnt = 0
	getAccountFromHash(accBHash).TxCnt = 0
	fundsStateChange(funds)
	if accA.Balance != balanceA || accB.Balance != balanceB {
		t.Error("State update failed!")
	}
	fundsStateChangeRollback(funds)
	if accA.Balance != rollBackA || accB.Balance != rollBackB {
		t.Error("Rollback failed!")
	}
	minerBal := minerAcc.Balance
	collectTxFees(funds,nil,minerAccHash)
	if feeA+feeB != minerAcc.Balance-minerBal {
		t.Error("Fee Collection failed!")
	}
	collectTxFeesRollback(funds,nil,minerAccHash)
	if minerBal != minerAcc.Balance {
		t.Error("Fee Collection Rollback failed!")
	}
	balBeforeRew := minerAcc.Balance
	reward := 5
	collectBlockReward(uint64(reward),minerAccHash)
	if minerAcc.Balance != balBeforeRew+uint64(reward) {
		t.Error("Block reward collection failed!")
	}
	collectBlockRewardRollback(uint64(reward),minerAccHash)
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
			singleAccHash := serializeHashContent(singleAcc.Address)
			if singleAccHash == accHash {
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
			singleAccHash := serializeHashContent(singleAcc.Address)
			if singleAccHash == accHash {
				found = true
			}
		}
		if found {
			t.Errorf("Account State failed to rollback the following account: %v\n", acc)
		}
	}
}

func TestCollectTxFeesRollback(t *testing.T) {

	rand := rand.New(rand.NewSource(time.Now().Unix()))

	var funds, funds2 []*fundsTx

	accAHash := serializeHashContent(accA.Address)
	accBHash := serializeHashContent(accB.Address)
	minerHash := serializeHashContent(minerAcc.Address)

	minerBal := minerAcc.Balance
	//rollback everything
	var fee uint64
	loopMax := int(rand.Uint64()%1000)
	for i := 0; i < loopMax+1; i++ {
		tx, _ := ConstrFundsTx(0x01,rand.Uint64()%1000000+1, rand.Uint64()%100+1, uint32(i), accAHash, accBHash, &PrivKeyA)

		funds = append(funds,tx)
		fee += tx.Fee
	}

	collectTxFees(funds,nil,minerHash)
	if minerBal+fee != minerAcc.Balance {
		t.Errorf("%v + %v != %v\n", minerBal,fee,minerAcc.Balance)
	}
	collectTxFeesRollback(funds,nil,minerHash)
	if minerBal != minerAcc.Balance {
		t.Errorf("Tx fees rollback failed: %v != %v\n", minerBal, minerAcc.Balance)
	}



	minerAcc.Balance = MAX_MONEY-100
	var fee2 uint64
	minerBal = minerAcc.Balance
	//interrupt somewhere in between
	for i := 2; i < 100; i++ {
		tx, _ := ConstrFundsTx(0x01,rand.Uint64()%1000000+1, uint64(i), uint32(i), accAHash, accBHash, &PrivKeyA)
		funds2 = append(funds2,tx)
		fee2 += tx.Fee
	}

	//should throw an error and result in a rollback, because of acc balance overflow
	if err := collectTxFees(funds2, nil,minerHash); err == nil || minerBal != minerAcc.Balance {
		t.Errorf("No rollback resulted, %v != %v\n", minerBal, minerAcc.Balance)
	}
}

func TestCollectBlockRewardRollback(t *testing.T) {

}