package bc

import (
	"testing"
	"math/rand"
)

func TestFundsTx(t *testing.T) {

	for i := 1; i < 10000; i++ {
		tx, _ := ConstrFundsTx(0x01,uint32(rand.Int31()%100000+1), uint16(rand.Uint32()), uint32(i), accA.Hash, accB.Hash, &PrivKeyA)
		if tx.verify() == false {
			t.Errorf("Tx could not be verified: \n%v", tx)
		}
	}
}

