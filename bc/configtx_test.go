package bc

import (
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func TestConfigTx(t *testing.T) {
	rand := rand.New(rand.NewSource(time.Now().Unix()))

	//creating some root-signed config txs
	tx, err := ConstrConfigTx(uint8(rand.Uint32()%256), 1, 5000, rand.Uint64(), &RootPrivKey)
	tx2, err2 := ConstrConfigTx(uint8(rand.Uint32()%256), 2, 5000, rand.Uint64(), &RootPrivKey)
	tx3, err3 := ConstrConfigTx(uint8(rand.Uint32()%256), 3, 5000, rand.Uint64(), &RootPrivKey)
	tx4, err4 := ConstrConfigTx(uint8(rand.Uint32()%256), 4, 5000, rand.Uint64(), &RootPrivKey)
	tx5, err5 := ConstrConfigTx(uint8(rand.Uint32()%256), 5, 5000, rand.Uint64(), &RootPrivKey)
	txfail, err6 := ConstrConfigTx(uint8(rand.Uint32()%256), 20, 5000, rand.Uint64(), &RootPrivKey)

	if (tx.verify() == false || err != nil) &&
		(tx2.verify() == false || err2 != nil) &&
		(tx3.verify() == false || err3 != nil) &&
		(tx4.verify() == false || err4 != nil) &&
		(tx5.verify() == false || err5 != nil) &&
		(txfail.verify() == true || err6 != nil) {
		t.Error("ConfigTx verification malfunctioning!")
	}
}

func TestConfigTxSerialization(t *testing.T) {
	rand := rand.New(rand.NewSource(time.Now().Unix()))

	loopMax := int(rand.Uint32() % 10000)
	for i := 0; i < loopMax; i++ {
		tx, err := ConstrConfigTx(uint8(rand.Uint32()%256), uint8(rand.Uint32()%256), rand.Uint64(), rand.Uint64(), &RootPrivKey)
		data := EncodeConfigTx(tx)
		decodedTx := DecodeConfigTx(data)
		if !reflect.DeepEqual(tx, decodedTx) || err != nil {
			t.Errorf("ConfigTx Serialization failed (%v) vs. (%v)\n", tx, decodedTx)
		}
	}
}
