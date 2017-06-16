package bc

import (
	"fmt"
	"reflect"
	"testing"
)

func TestValidateBlockRollback(t *testing.T) {

	cleanAndPrepare()
	b := newBlock()

	accsBefore := make(map[[64]byte]Account)
	accsBefore2 := make(map[[64]byte]Account)
	accsAfter := make(map[[64]byte]Account)

	for _, accSlice := range State {
		for _, acc := range accSlice {
			accsBefore[acc.Address] = *acc
		}
	}

	createBlockWithTxs(b)
	b.finalizeBlock()
	validateBlock(b)

	for _, accSlice := range State {
		for _, acc := range accSlice {
			accsAfter[acc.Address] = *acc
		}
	}

	if reflect.DeepEqual(accsBefore, accsAfter) {
		t.Error("State wasn't changed despite validating a block!")
	}

	validateBlockRollback(b)

	for _, accSlice := range State {
		for _, acc := range accSlice {
			accsBefore2[acc.Address] = *acc
		}
	}

	if !reflect.DeepEqual(accsBefore, accsBefore2) {
		t.Error("State wasn't rolled back")
	}
}

func TestMultipleBlocksRollback(t *testing.T) {
	//create 4 blocks after genesis, rollback 3
	cleanAndPrepare()

	stategenesis := make(map[[64]byte]Account)
	stateb := make(map[[64]byte]Account)
	stateb2 := make(map[[64]byte]Account)
	stateb3 := make(map[[64]byte]Account)
	tmpState := make(map[[64]byte]Account)

	//system parameters
	var paramgenesis []parameters
	var paramb []parameters
	var paramb2 []parameters
	var paramb3 []parameters

	//no deep copy, becasue we use []*Account
	for _, accSlice := range State {
		for _, acc := range accSlice {
			stategenesis[acc.Address] = *acc
		}
	}
	paramgenesis = make([]parameters, len(parameterSlice))
	copy(paramgenesis, parameterSlice)

	b := newBlock()
	createBlockWithTxs(b)
	b.finalizeBlock()
	if err := validateBlock(b); err != nil {
		t.Errorf("Block validation for (%v) failed: %v\n", b, err)
	}

	for _, accSlice := range State {
		for _, acc := range accSlice {
			stateb[acc.Address] = *acc
		}
	}
	paramb = make([]parameters, len(parameterSlice))
	copy(paramb, parameterSlice)

	b2 := newBlock()
	b2.PrevHash = b.Hash
	createBlockWithTxs(b2)
	b2.finalizeBlock()
	if err := validateBlock(b2); err != nil {
		t.Errorf("Block failed: %v\n", b2)
	}

	for _, accSlice := range State {
		for _, acc := range accSlice {
			stateb2[acc.Address] = *acc
		}
	}
	paramb2 = make([]parameters, len(parameterSlice))
	copy(paramb2, parameterSlice)

	b3 := newBlock()
	b3.PrevHash = b2.Hash
	createBlockWithTxs(b3)
	b3.finalizeBlock()
	if err := validateBlock(b3); err != nil {
		t.Errorf("Block failed: %v\n", b3)
	}

	for _, accSlice := range State {
		for _, acc := range accSlice {
			stateb3[acc.Address] = *acc
		}
	}
	paramb3 = make([]parameters, len(parameterSlice))
	copy(paramb3, parameterSlice)

	b4 := newBlock()
	b4.PrevHash = b3.Hash
	createBlockWithTxs(b4)
	b4.finalizeBlock()
	if err := validateBlock(b4); err != nil {
		t.Errorf("Block failed: %v\n", b4)
	}

	//STARTING ROLLBACKS---------------------------------------------
	if err := validateBlockRollback(b4); err != nil {
		t.Errorf("%v\n", err)
	}
	for _, accSlice := range State {
		for _, acc := range accSlice {
			tmpState[acc.Address] = *acc
		}
	}

	if !reflect.DeepEqual(tmpState, stateb3) || !reflect.DeepEqual(paramb3, parameterSlice) {
		t.Error("Block rollback failed.")
	}
	//delete tmpState
	for k := range tmpState {
		delete(tmpState, k)
	}

	if err := validateBlockRollback(b3); err != nil {
		t.Errorf("%v\n", err)
	}
	for _, accSlice := range State {
		for _, acc := range accSlice {
			tmpState[acc.Address] = *acc
		}
	}
	if !reflect.DeepEqual(tmpState, stateb2) || !reflect.DeepEqual(paramb2, parameterSlice) {

		for _, entry := range paramb2 {
			fmt.Printf("%v\n", entry)
		}
		fmt.Println()
		for _, entry := range parameterSlice {
			fmt.Printf("%v\n", entry)
		}

		t.Error("Block rollback failed.")
	}
	for k := range tmpState {
		delete(tmpState, k)
	}

	if err := validateBlockRollback(b2); err != nil {
		t.Errorf("%v\n", err)
	}
	for _, accSlice := range State {
		for _, acc := range accSlice {
			tmpState[acc.Address] = *acc
		}
	}
	if !reflect.DeepEqual(tmpState, stateb) || !reflect.DeepEqual(paramb, parameterSlice) {
		t.Error("Block rollback failed.")
	}
	for k := range tmpState {
		delete(tmpState, k)
	}

	if err := validateBlockRollback(b); err != nil {
		t.Errorf("%v\n", err)
	}
	for _, accSlice := range State {
		for _, acc := range accSlice {
			tmpState[acc.Address] = *acc
		}
	}
	if !reflect.DeepEqual(tmpState, stategenesis) || !reflect.DeepEqual(paramgenesis, parameterSlice) {
		t.Error("Block rollback failed.")
	}
	for k := range tmpState {
		delete(tmpState, k)
	}
}
