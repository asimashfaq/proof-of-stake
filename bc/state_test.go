package bc

import (
	"math/rand"
	"testing"
	"time"
)

//Testing state change, rollback and fee collection
func TestFundsTxStateChange(t *testing.T) {

	rand := rand.New(rand.NewSource(time.Now().Unix()))

	accAHash := serializeHashContent(accA.Address)
	accBHash := serializeHashContent(accB.Address)
	minerAccHash := serializeHashContent(minerAcc.Address)

	var testSize uint32
	testSize = 1000

	b := newBlock()
	var funds []*fundsTx

	var feeA, feeB uint64

	//we're testing an overflowing balance in another test, this is that no interference occurs
	accA.Balance = 12343478374563434
	accB.Balance = 2947939489348234
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

		ftx2,_ := ConstrFundsTx(0x01,rand.Uint64()%1000+1, rand.Uint64()%100+1, uint32(i), accAHash, accAHash, &PrivKeyB)
		if b.addTx(ftx2) == nil {
			funds = append(funds,ftx2)
			balanceB -= ftx2.Amount
			feeB += ftx2.Fee

			balanceA += ftx2.Amount
		}
	}

	getAccountFromHash(accAHash).TxCnt = 0
	getAccountFromHash(accBHash).TxCnt = 0


	//we have another test that checks overflows, have to set the balance to a fixed lower value
	minerAcc.Balance = 123456
	minerBal := minerAcc.Balance
	fundsStateChange(funds)

	if accA.Balance != balanceA || accB.Balance != balanceB {
		t.Errorf("State update failed: %v != %v or %v != %v\n", accA.Balance, balanceA, accB.Balance, balanceB)
	}


	collectTxFees(funds,nil,nil,minerAccHash)
	if feeA+feeB != minerAcc.Balance-minerBal {
		t.Error("Fee Collection failed!")
	}

	balBeforeRew := minerAcc.Balance
	collectBlockReward(BLOCK_REWARD,minerAccHash)
	if minerAcc.Balance != balBeforeRew+BLOCK_REWARD {
		t.Error("Block reward collection failed!")
	}
}

func TestAccountOverflow(t *testing.T) {

	var accSlice []*fundsTx
	accAHash := serializeHashContent(accA.Address)
	accBHash := serializeHashContent(accB.Address)

	accA.Balance = MAX_MONEY
	accA.TxCnt = 0
	tx,err := ConstrFundsTx(0x01, 1,1,0, accBHash, accAHash, &PrivKeyB)
	if !tx.verify() || err != nil {
		t.Error("Failed to create reasonable fundsTx\n")
		return
	}
	accSlice = append(accSlice,tx)
	err = fundsStateChange(accSlice)

	//err shouldn't be nil, because the tx can't have been successful
	//also, the balance of A shouldn't have changed
	if err == nil || accA.Balance != MAX_MONEY {
		t.Errorf("Failed to block overflowing transaction to account with balance: %v\n", accA.Balance)
	}
}

func TestAccTxStateChange(t *testing.T) {

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
		//make sure the previously created acc is in the state
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
}

func TestConfigTxStateChange(t *testing.T) {

	rand := rand.New(rand.NewSource(time.Now().Unix()))
	var testSize uint32
	testSize = 1000
	var accs []*accTx

	loopMax := int(rand.Uint32()%testSize)+1
	for i := 0; i < loopMax; i++ {
		tx,_ := ConstrAccTx(rand.Uint64()%1000,&RootPrivKey)
		accs = append(accs, tx)
	}


}