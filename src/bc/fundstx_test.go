package bc

import (
	"testing"
	"math/rand"
	"time"
)

func TestFundsTx(t *testing.T) {
	rand := rand.New(rand.NewSource(time.Now().Unix()))

	for i := 1; i < 10000; i++ {
		tx, _ := ConstrFundsTx(0x01,rand.Uint64()%100000+1, rand.Uint64()%10+1, uint32(i), accA.Hash, accB.Hash, &PrivKeyA)
		if tx.verify() == false {
			t.Errorf("Tx could not be verified: \n%v", tx)
		}
	}
}

