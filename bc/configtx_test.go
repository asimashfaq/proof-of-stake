package bc

import (
	"testing"
	"time"
	"reflect"
	"math/rand"
)

func TestConfigTx(t *testing.T) {
	rand := rand.New(rand.NewSource(time.Now().Unix()))

	//creating some root-signed config txs
	loopMax := int(rand.Uint64()%1000)
	for i := 0; i <= loopMax; i++ {
		tx,err := ConstrConfigTx(uint8(rand.Uint32()%256), uint8(rand.Uint32()%256),rand.Uint64(), rand.Uint64(), &RootPrivKey)
		if tx.verify() == false || err != nil {
			t.Errorf("ConfigTx could not be verified: %v\n", tx)
		}
	}
}

func TestConfigTxSerialization(t *testing.T) {
	rand := rand.New(rand.NewSource(time.Now().Unix()))

	loopMax := int(rand.Uint32()%10000)
	for i := 0; i < loopMax; i++ {
		tx,err := ConstrConfigTx(uint8(rand.Uint32()%256), uint8(rand.Uint32()%256),rand.Uint64(), rand.Uint64(), &RootPrivKey)
		data := EncodeConfigTx(tx)
		decodedTx := DecodeConfigTx(data)
		if !reflect.DeepEqual(tx, decodedTx) || err != nil {
			t.Errorf("ConfigTx Serialization failed (%v) vs. (%v)\n", tx, decodedTx)
		}
	}
}
