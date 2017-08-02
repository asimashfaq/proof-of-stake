package protocol

import (
	"testing"
	"math/rand"
	"time"
	"reflect"
	"fmt"
)

func TestBlockSerialization(t *testing.T) {
	rand := rand.New(rand.NewSource(time.Now().Unix()))

	b := new(Block)
	b.Hash = [32]byte{0,1,2,3,4}
	b.PrevHash = [32]byte{1,2,3,4,5}
	b.Header = byte(rand.Int31())
	b.Nonce = [8]byte{0,1,2,3,4,5,6,7}
	b.Timestamp = time.Now().Unix()
	b.MerkleRoot = [32]byte{2,3,4,5,6}
	b.Beneficiary = [32]byte{3,4,5,6,7}
	b.NrAccTx = uint16(rand.Uint32())
	b.NrFundsTx = uint16(rand.Uint32())
	b.NrConfigTx = uint8(rand.Uint32())

	encodedBlock := b.Encode()
	b2 := b.Decode(encodedBlock)

	if !reflect.DeepEqual(encodedBlock,b2.Encode()) {
		t.Error("Block encoding/decoding failed\n")
	}
}

func TestGetSize(t *testing.T) {
	b := new(Block)

	b.NrAccTx = uint16(rand.Uint32())
	b.NrFundsTx = uint16(rand.Uint32())
	b.NrConfigTx = uint8(rand.Uint32())

	txAmount := b.NrAccTx + b.NrFundsTx + uint16(b.NrConfigTx)

	if b.GetSize() != uint64(txAmount) * 32 + BLOCKHEADER_SIZE {
		fmt.Printf("Miscalculated block size: %v vs. %v\n", b.GetSize(), uint64(txAmount) * 32 + BLOCKHEADER_SIZE)
	}
}

