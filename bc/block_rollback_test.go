package bc

import (
	"testing"
	"reflect"
)

func TestValidateBlockRollback(t *testing.T) {

	cleanAndPrepare()

	//this is our pre-block state
	var accs []Account
	for _,accSlice := range State {
		for _,acc := range accSlice {
			accs = append(accs,*acc)
		}
	}

	b := newBlock()
	createBlockWithTxs(b)
	b.finalizeBlock()

	validateBlock(b)

	//maybe also add the Open/ClosedTx memory check for postValidationRollback
	if len(State) == len(accs) {
		t.Error("Block validation failed!\n")
	}

	validateBlockRollback(b)
	//we need to have the same acc state as before
	if len(State) != len(accs) {
		t.Errorf("Rollback failed: len(State) = %v vs. len(accs) = %v\n", len(State), len(accs))
	}

	for _,acc := range accs {
		accHash := serializeHashContent(acc.Address)
		stateAcc := getAccountFromHash(accHash)
		if !reflect.DeepEqual(*stateAcc,acc) {
			t.Errorf("The following accounts were not the same after the rollback\n%v\n\nvs.\n\n%v\n", stateAcc,acc)
		}
	}
}