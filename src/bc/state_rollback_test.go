package bc

import (
	"testing"
	"time"
	"math/rand"
)

func TestFundsStateChangeRollback(t *testing.T) {

	rand := rand.New(rand.NewSource(time.Now().Unix()))

	var txSlice []*fundsTx
	//creating some root-signed new accounts
	loopMax := int(rand.Uint32()%10000)
	for i := 0; i < loopMax; i++ {
		tx, _ := ConstrFundsTx(0x01,rand.Uint64()%100000+1, rand.Uint64()%10+1, uint32(i), accA.Hash, accB.Hash, &PrivKeyA)
		txSlice = append(txSlice,tx)
		if tx.verify() == false {
			t.Errorf("Tx could not be verified: \n%v", tx)
		}
	}

	fundsStateChange(txSlice)
}