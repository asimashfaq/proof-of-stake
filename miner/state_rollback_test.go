package miner

import (
	"github.com/lisgie/bazo_miner/protocol"
	"github.com/lisgie/bazo_miner/storage"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func TestFundsStateChangeRollback(t *testing.T) {

	cleanAndPrepare()
	rand := rand.New(rand.NewSource(time.Now().Unix()))

	accAHash := serializeHashContent(accA.Address)
	accBHash := serializeHashContent(accB.Address)
	minerAccHash := serializeHashContent(minerAcc.Address)

	var testSize uint32
	testSize = 1000

	b := newBlock([32]byte{})
	var funds []*protocol.FundsTx

	var feeA, feeB uint64

	rollBackA := accA.Balance
	rollBackB := accB.Balance

	balanceA := accA.Balance
	balanceB := accB.Balance

	loopMax := int(rand.Uint32()%testSize + 1)
	for i := 0; i < loopMax+1; i++ {
		ftx, _ := protocol.ConstrFundsTx(0x01, rand.Uint64()%1000000+1, rand.Uint64()%100+1, uint32(i), accAHash, accBHash, &PrivKeyA)
		if addTx(b, ftx) == nil {
			funds = append(funds, ftx)
			balanceA -= ftx.Amount
			feeA += ftx.Fee

			balanceB += ftx.Amount
		} else {
			t.Errorf("Block rejected a valid transaction: %v\n", ftx)
		}

		ftx2, _ := protocol.ConstrFundsTx(0x01, rand.Uint64()%1000+1, rand.Uint64()%100+1, uint32(i), accBHash, accAHash, &PrivKeyB)
		if addTx(b, ftx2) == nil {
			funds = append(funds, ftx2)
			balanceB -= ftx2.Amount
			feeB += ftx2.Fee

			balanceA += ftx2.Amount
		} else {
			t.Errorf("Block rejected a valid transaction: %v\n", ftx2)
		}
	}
	fundsStateChange(funds)
	if accA.Balance != balanceA || accB.Balance != balanceB {
		t.Error("State update failed!")
	}
	fundsStateChangeRollback(funds)
	if accA.Balance != rollBackA || accB.Balance != rollBackB {
		t.Error("Rollback failed!")
	}

	//collectTxFees is checked below in its own test (to additionally cover overflow scenario)
	balBeforeRew := minerAcc.Balance
	reward := 5
	collectBlockReward(uint64(reward), minerAccHash)
	if minerAcc.Balance != balBeforeRew+uint64(reward) {
		t.Error("Block reward collection failed!")
	}
	collectBlockRewardRollback(uint64(reward), minerAccHash)
	if minerAcc.Balance != balBeforeRew {
		t.Error("Block reward collection rollback failed!")
	}
}

func TestAccStateChangeRollback(t *testing.T) {

	cleanAndPrepare()
	rand := rand.New(rand.NewSource(time.Now().Unix()))

	var testSize uint32
	testSize = 1000

	var accs []*protocol.AccTx

	loopMax := int(rand.Uint32()%testSize) + 1
	for i := 0; i < loopMax; i++ {
		tx, _ := protocol.ConstrAccTx(0, rand.Uint64()%1000, &RootPrivKey)
		accs = append(accs, tx)
	}

	accStateChange(accs)

	var shortHash [8]byte
	for _, acc := range accs {
		found := false
		accHash := serializeHashContent(acc.PubKey)
		copy(shortHash[:], accHash[0:8])
		accSlice := storage.State[shortHash]
		for _, singleAcc := range accSlice {
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

	for _, acc := range accs {
		found := false
		accHash := serializeHashContent(acc.PubKey)
		copy(shortHash[:], accHash[0:8])
		accSlice := storage.State[shortHash]
		for _, singleAcc := range accSlice {
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

func TestConfigStateChangeRollback(t *testing.T) {
	cleanAndPrepare()

	var configSlice []*protocol.ConfigTx

	tx, _ := protocol.ConstrConfigTx(uint8(rand.Uint32()%256), 1, 1000, rand.Uint64(), &RootPrivKey)
	tx2, _ := protocol.ConstrConfigTx(uint8(rand.Uint32()%256), 2, 2000, rand.Uint64(), &RootPrivKey)
	tx3, _ := protocol.ConstrConfigTx(uint8(rand.Uint32()%256), 3, 3000, rand.Uint64(), &RootPrivKey)
	tx4, _ := protocol.ConstrConfigTx(uint8(rand.Uint32()%256), 4, 4000, rand.Uint64(), &RootPrivKey)
	tx5, _ := protocol.ConstrConfigTx(uint8(rand.Uint32()%256), 5, 5000, rand.Uint64(), &RootPrivKey)

	configSlice = append(configSlice, tx)
	configSlice = append(configSlice, tx2)
	configSlice = append(configSlice, tx3)
	configSlice = append(configSlice, tx4)
	configSlice = append(configSlice, tx5)

	before := *activeParameters
	configStateChange(configSlice, [32]byte{'0', '1', '2'})
	if reflect.DeepEqual(before, *activeParameters) {
		t.Error("No config state change.")
	}
	configStateChangeRollback(configSlice, [32]byte{'0', '1', '2'})
	if !reflect.DeepEqual(before, *activeParameters) {
		t.Error("Config state rollback failed.")
	}
}

func TestCollectTxFeesRollback(t *testing.T) {

	cleanAndPrepare()
	rand := rand.New(rand.NewSource(time.Now().Unix()))

	var funds, funds2 []*protocol.FundsTx

	accAHash := serializeHashContent(accA.Address)
	accBHash := serializeHashContent(accB.Address)
	minerHash := serializeHashContent(minerAcc.Address)

	minerBal := minerAcc.Balance
	//rollback everything
	var fee uint64
	loopMax := int(rand.Uint64() % 1000)
	for i := 0; i < loopMax+1; i++ {
		tx, _ := protocol.ConstrFundsTx(0x01, rand.Uint64()%1000000+1, rand.Uint64()%100+1, uint32(i), accAHash, accBHash, &PrivKeyA)

		funds = append(funds, tx)
		fee += tx.Fee
	}

	collectTxFees(funds, nil, nil, minerHash)
	if minerBal+fee != minerAcc.Balance {
		t.Errorf("%v + %v != %v\n", minerBal, fee, minerAcc.Balance)
	}
	collectTxFeesRollback(funds, nil, nil, minerHash)
	if minerBal != minerAcc.Balance {
		t.Errorf("Tx fees rollback failed: %v != %v\n", minerBal, minerAcc.Balance)
	}

	minerAcc.Balance = protocol.MAX_MONEY - 100
	var fee2 uint64
	minerBal = minerAcc.Balance
	//interrupt somewhere in between
	for i := 2; i < 100; i++ {
		tx, _ := protocol.ConstrFundsTx(0x01, rand.Uint64()%1000000+1, uint64(i), uint32(i), accAHash, accBHash, &PrivKeyA)
		funds2 = append(funds2, tx)
		fee2 += tx.Fee
	}

	accABal := accA.Balance
	accBBal := accB.Balance
	//should throw an error and result in a rollback, because of acc balance overflow
	tmpBlock := newBlock([32]byte{})
	tmpBlock.Beneficiary = minerHash
	data := blockData{funds2, nil, nil, tmpBlock}
	if err := stateValidation(data); err == nil ||
		minerBal != minerAcc.Balance ||
		accA.Balance != accABal ||
		accB.Balance != accBBal {
		t.Errorf("No rollback resulted, %v != %v\n", minerBal, minerAcc.Balance)
	}
}
