package bc

import (
	"testing"
	"math/rand"
	"time"
	"reflect"
)

func TestFundsTxVerification(t *testing.T) {
	rand := rand.New(rand.NewSource(time.Now().Unix()))

	for i := 1; i < int(rand.Uint32()%10000); i++ {
		tx, _ := ConstrFundsTx(0x01,rand.Uint64()%100000+1, rand.Uint64()%10+1, uint32(i), accA.Hash, accB.Hash, &PrivKeyA)
		if tx.verify() == false {
			t.Errorf("Tx could not be verified: \n%v", tx)
		}
	}
}

func TestFundsTxSerialization(t *testing.T) {
	rand := rand.New(rand.NewSource(time.Now().Unix()))

	for i := 1; i < int(rand.Uint32()%10000); i++ {
		tx, _ := ConstrFundsTx(0x01,rand.Uint64()%100000+1, rand.Uint64()%10+1, uint32(i), accA.Hash, accB.Hash, &PrivKeyA)
		data := EncodeFundsTx(tx)
		decodedTx := DecodeFundsTx(data)
		if !reflect.DeepEqual(tx, decodedTx) {
			t.Errorf("FundsTx Serialization failed (%v) vs. (%v)\n", tx, decodedTx)
		}
	}
}