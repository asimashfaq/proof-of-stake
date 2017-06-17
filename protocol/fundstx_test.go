package protocol

import (
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func TestFundsTxVerification(t *testing.T) {
	rand := rand.New(rand.NewSource(time.Now().Unix()))
	accAHash := serializeHashContent(accA.Address)
	accBHash := serializeHashContent(accB.Address)
	loopMax := int(rand.Uint32() % 10000)
	for i := 0; i < loopMax; i++ {
		tx, _ := ConstrFundsTx(0x01, rand.Uint64()%100000+1, rand.Uint64()%10+1, uint32(i), accAHash, accBHash, &PrivKeyA)
		if tx.verify() == false {
			t.Errorf("Tx could not be verified: \n%v", tx)
		}
	}
}

func TestFundsTxSerialization(t *testing.T) {
	rand := rand.New(rand.NewSource(time.Now().Unix()))
	accAHash := serializeHashContent(accA.Address)
	accBHash := serializeHashContent(accB.Address)
	loopMax := int(rand.Uint32() % 10000)
	for i := 0; i < loopMax; i++ {
		tx, _ := ConstrFundsTx(0x01, rand.Uint64()%100000+1, rand.Uint64()%10+1, uint32(i), accAHash, accBHash, &PrivKeyA)
		data := EncodeFundsTx(tx)
		decodedTx := DecodeFundsTx(data)
		decodedTx.verify()
		if !reflect.DeepEqual(tx, decodedTx) {
			t.Errorf("FundsTx Serialization failed (%v) vs. (%v)\n", tx, decodedTx)
		}
	}
}
