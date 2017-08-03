package miner

import (
	"github.com/lisgie/bazo_miner/protocol"
	"github.com/lisgie/bazo_miner/storage"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

//Testing state change, rollback and fee collection
func TestFundsTxStateChange(t *testing.T) {

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

	//we're testing an overflowing balance in another test, this is that no interference occurs
	accA.Balance = 12343478374563434
	accB.Balance = 2947939489348234
	balanceA := accA.Balance
	balanceB := accB.Balance
	minerBal := minerAcc.Balance

	loopMax := int(rand.Uint32()%testSize + 1)
	for i := 0; i < loopMax+1; i++ {
		ftx, _ := protocol.ConstrFundsTx(0x01, rand.Uint64()%1000000+1, rand.Uint64()%100+1, uint32(i), accAHash, accBHash, &PrivKeyA)
		if addTx(b, ftx) == nil {
			funds = append(funds, ftx)
			balanceA -= ftx.Amount
			feeA += ftx.Fee

			balanceB += ftx.Amount
		}

		ftx2, _ := protocol.ConstrFundsTx(0x01, rand.Uint64()%1000+1, rand.Uint64()%100+1, uint32(i), accAHash, accAHash, &PrivKeyB)
		if addTx(b, ftx2) == nil {
			funds = append(funds, ftx2)
			balanceB -= ftx2.Amount
			feeB += ftx2.Fee

			balanceA += ftx2.Amount
		}
	}

	fundsStateChange(funds)

	if accA.Balance != balanceA || accB.Balance != balanceB {
		t.Errorf("State update failed: %v != %v or %v != %v\n", accA.Balance, balanceA, accB.Balance, balanceB)
	}

	collectTxFees(nil, funds, nil, minerAccHash)
	if feeA+feeB != minerAcc.Balance-minerBal {
		t.Error("Fee Collection failed!")
	}

	balBeforeRew := minerAcc.Balance
	collectBlockReward(activeParameters.block_reward, minerAccHash)
	if minerAcc.Balance != balBeforeRew+activeParameters.block_reward {
		t.Error("Block reward collection failed!")
	}
}

func TestAccountOverflow(t *testing.T) {

	cleanAndPrepare()
	var accSlice []*protocol.FundsTx
	accAHash := serializeHashContent(accA.Address)
	accBHash := serializeHashContent(accB.Address)

	accA.Balance = MAX_MONEY
	accA.TxCnt = 0
	tx, err := protocol.ConstrFundsTx(0x01, 1, 1, 0, accBHash, accAHash, &PrivKeyB)
	if !verifyFundsTx(tx) || err != nil {
		t.Error("Failed to create reasonable fundsTx\n")
		return
	}
	accSlice = append(accSlice, tx)
	err = fundsStateChange(accSlice)

	//Err shouldn't be nil, because the tx can't have been successful
	//Also, the balance of A shouldn't have changed
	if err == nil || accA.Balance != MAX_MONEY {
		t.Errorf("Failed to block overflowing transaction to account with balance: %v\n", accA.Balance)
	}
}

func TestAccTxStateChange(t *testing.T) {

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

	for _, acc := range accs {
		accHash := serializeHashContent(acc.PubKey)
		acc := storage.State[accHash]
		//make sure the previously created acc is in the state
		if acc == nil {
			t.Errorf("Account State failed to update for the following account: %v\n", acc)
		}
	}

	//Create a new root account, set the header to 0x01
	var singleSlice []*protocol.AccTx
	tx, _ := protocol.ConstrAccTx(0x01, rand.Uint64()%1000, &RootPrivKey)
	singleSlice = append(singleSlice, tx)
	var pubKeyTmp [64]byte
	copy(pubKeyTmp[:], tx.PubKey[:])

	accStateChange(singleSlice)

	if !isRootKey(serializeHashContent(pubKeyTmp)) {
		t.Errorf("AccTx Header bit 1 not working.")
	}

	//Set header to 0x02 -> delete root account
	newTx := *tx
	newTx.Header = 0x02
	singleSlice[0] = &newTx
	accStateChange(singleSlice)

	if isRootKey(serializeHashContent(pubKeyTmp)) {
		t.Errorf("AccTx Header bit 2 not working.")
	}
}

func TestConfigTxStateChange(t *testing.T) {

	rand := rand.New(rand.NewSource(time.Now().Unix()))
	var testSize uint32
	testSize = 1000
	var configs []*protocol.ConfigTx

	loopMax := int(rand.Uint32()%testSize) + 1
	for i := 0; i < loopMax; i++ {
		tx, err := protocol.ConstrConfigTx(uint8(rand.Uint32()%256), uint8(rand.Uint32()%5+1), rand.Uint64()%10000000, rand.Uint64(), uint8(i), &RootPrivKey)
		if err != nil {
			t.Errorf("ConfigTx Creation failed (%v)\n", err)
		}
		if verifyConfigTx(tx) {
			configs = append(configs, tx)
		}
	}
	parameterSet := *activeParameters
	tmpLen := len(parameterSlice)
	configStateChange(configs, [32]byte{'0', '1'})
	parameterSet2 := *activeParameters
	if tmpLen != len(parameterSlice)-1 || reflect.DeepEqual(parameterSet, parameterSet2) {
		t.Errorf("Config State Change malfunctioned: %v != %v\n", tmpLen, len(parameterSlice)-1)
	}

	cleanAndPrepare()
	var configs2 []*protocol.ConfigTx
	//test the inner workings of configStateChange as well...
	tx, _ := protocol.ConstrConfigTx(uint8(rand.Uint32()%256), 1, 1000, rand.Uint64(), 0, &RootPrivKey)
	tx2, _ := protocol.ConstrConfigTx(uint8(rand.Uint32()%256), 2, 2000, rand.Uint64(), 0, &RootPrivKey)
	tx3, _ := protocol.ConstrConfigTx(uint8(rand.Uint32()%256), 3, 3000, rand.Uint64(), 0, &RootPrivKey)
	tx4, _ := protocol.ConstrConfigTx(uint8(rand.Uint32()%256), 4, 4000, rand.Uint64(), 0, &RootPrivKey)
	tx5, _ := protocol.ConstrConfigTx(uint8(rand.Uint32()%256), 5, 5000, rand.Uint64(), 0, &RootPrivKey)

	configs2 = append(configs2, tx)
	configs2 = append(configs2, tx2)
	configs2 = append(configs2, tx3)
	configs2 = append(configs2, tx4)
	configs2 = append(configs2, tx5)

	configStateChange(configs2, [32]byte{})
	if activeParameters.block_size != 1000 ||
		activeParameters.diff_interval != 2000 ||
		activeParameters.fee_minimum != 3000 ||
		activeParameters.block_interval != 4000 ||
		activeParameters.block_reward != 5000 {
		t.Error("Config StateChanged didn't set the correct parameters!")
	}
}

//If we parse configTxs which are unknown, we don't change parameter datastructure
func TestConfigTxStateChangeUnknown(t *testing.T) {

	cleanAndPrepare()
	//Issuing configTxs with unknown Id
	var configs []*protocol.ConfigTx
	tx, _ := protocol.ConstrConfigTx(uint8(rand.Uint32()%256), 11, 1000, rand.Uint64(), 0, &RootPrivKey)
	tx2, _ := protocol.ConstrConfigTx(uint8(rand.Uint32()%256), 11, 2000, rand.Uint64(), 0, &RootPrivKey)
	tx3, _ := protocol.ConstrConfigTx(uint8(rand.Uint32()%256), 11, 3000, rand.Uint64(), 0, &RootPrivKey)

	//save parameter state
	tmpParameter := parameterSlice[len(parameterSlice)-1]

	configs = append(configs, tx)
	configs = append(configs, tx2)
	configs = append(configs, tx3)

	configStateChange(configs, [32]byte{'0', '1'})

	if !reflect.DeepEqual(tmpParameter, *activeParameters) {
		t.Error("Parameter state changed even though it shouldn't have.")
	}

	configStateChangeRollback(configs, [32]byte{'0', '1'})

	if !reflect.DeepEqual(tmpParameter, *activeParameters) {
		t.Error("Parameter state changed even though it shouldn't have.")
	}

	//Adding a tx that changes state
	tx4, _ := protocol.ConstrConfigTx(uint8(rand.Uint32()%256), 2, 3000, rand.Uint64(), 0, &RootPrivKey)
	configs = append(configs, tx4)

	configStateChange(configs, [32]byte{'0', '1'})

	if reflect.DeepEqual(tmpParameter, *activeParameters) {
		t.Error("Parameter state changed even though it shouldn't have.")
	}

	configStateChangeRollback(configs, [32]byte{'0', '1'})

	if !reflect.DeepEqual(tmpParameter, *activeParameters) {
		t.Error("Parameter state changed even though it shouldn't have.")
	}

	configStateChange(configs, [32]byte{'0', '1'})
	configStateChangeRollback(configs, [32]byte{'0'})
	//Only change if block hashes match
	if reflect.DeepEqual(tmpParameter, *activeParameters) {
		t.Error("Parameter state changed even though it shouldn't have.")
	}
}
