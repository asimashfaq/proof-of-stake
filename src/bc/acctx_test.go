package bc

import (
	"testing"
	"math/rand"
	"reflect"
	"time"
)

func TestAccTx(t *testing.T) {
	rand := rand.New(rand.NewSource(time.Now().Unix()))

	//creating some root-signed new accounts
	loopMax := int(rand.Uint64()%1000)
	for i := 0; i <= loopMax; i++ {
		tx, _ := ConstrAccTx(rand.Uint64()%100+1, &RootPrivKey)
		if tx.verify() == false {
			t.Errorf("AccTx could not be verified: %v\n", tx)
		}
	}
}

func TestAccTxSerialization(t *testing.T) {
	rand := rand.New(rand.NewSource(time.Now().Unix()))

	loopMax := int(rand.Uint32()%10000)
	for i := 1; i < loopMax; i++ {
		tx, _ := ConstrAccTx(rand.Uint64()%100+1, &RootPrivKey)
		data := EncodeAccTx(tx)
		decodedTx := DecodeAccTx(data)
		if !reflect.DeepEqual(tx, decodedTx) {
			t.Errorf("AccTx Serialization failed (%v) vs. (%v)\n", tx, decodedTx)
		}
	}
}