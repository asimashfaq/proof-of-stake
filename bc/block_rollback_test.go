package bc

import (
	"testing"
	"reflect"
)

func TestValidateBlockRollback(t *testing.T) {

	cleanAndPrepare()
	b := newBlock()

	var stateCopy map[[8]byte][]*Account
	stateCopy = make(map[[8]byte][]*Account)

	for k, v := range State {
		stateCopy[k] = v
	}

	createBlockWithTxs(b)
	b.finalizeBlock()
	validateBlock(b)

	if reflect.DeepEqual(stateCopy, State) {
		t.Error("State wasn't changed despite validating a block!")
	}

	validateBlockRollback(b)

	if !reflect.DeepEqual(stateCopy, State) {
		t.Error("State wasn't rolled back")
	}
}

func TestMultipleBlocksRollback(t *testing.T) {

}