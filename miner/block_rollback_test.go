package miner

import (
	"github.com/lisgie/bazo_miner/protocol"
	"github.com/lisgie/bazo_miner/storage"
	"reflect"
	"testing"
)

func TestValidateBlockRollback(t *testing.T) {

	cleanAndPrepare()
	b := newBlock([32]byte{})

	accsBefore := make(map[[64]byte]protocol.Account)
	accsBefore2 := make(map[[64]byte]protocol.Account)
	accsAfter := make(map[[64]byte]protocol.Account)

	for _, acc := range storage.State {
		accsBefore[acc.Address] = *acc
	}

	createBlockWithTxs(b)
	finalizeBlock(b)
	validateBlock(b)

	for _, acc := range storage.State {
		accsAfter[acc.Address] = *acc
	}

	if reflect.DeepEqual(accsBefore, accsAfter) {
		t.Error("State wasn't changed despite validating a block!")
	}

	err := validateBlockRollback(b)
	if err != nil {
		t.Errorf("%v\n", err)
	}

	for _, acc := range storage.State {
		accsBefore2[acc.Address] = *acc
	}

	if !reflect.DeepEqual(accsBefore, accsBefore2) {
		t.Error("State wasn't rolled back")
	}
}

func TestMultipleBlocksRollback(t *testing.T) {
	//create 4 blocks after genesis, rollback 3
	cleanAndPrepare()

	stateb := make(map[[64]byte]protocol.Account)
	stateb2 := make(map[[64]byte]protocol.Account)
	stateb3 := make(map[[64]byte]protocol.Account)
	tmpState := make(map[[64]byte]protocol.Account)

	//system parameters
	var paramb []parameters
	var paramb2 []parameters
	var paramb3 []parameters

	b := newBlock([32]byte{})
	createBlockWithTxs(b)
	finalizeBlock(b)
	if err := validateBlock(b); err != nil {
		t.Errorf("Block validation for (%v) failed: %v\n", b, err)
	}

	for _, acc := range storage.State {
		stateb[acc.Address] = *acc
	}

	paramb = make([]parameters, len(parameterSlice))
	copy(paramb, parameterSlice)

	b2 := newBlock(b.Hash)
	createBlockWithTxs(b2)
	finalizeBlock(b2)
	if err := validateBlock(b2); err != nil {
		t.Errorf("Block failed: %v\n", b2)
	}

	for _, acc := range storage.State {
		stateb2[acc.Address] = *acc
	}

	paramb2 = make([]parameters, len(parameterSlice))
	copy(paramb2, parameterSlice)

	b3 := newBlock(b2.Hash)
	createBlockWithTxs(b3)
	finalizeBlock(b3)
	if err := validateBlock(b3); err != nil {
		t.Errorf("Block failed: %v\n", b3)
	}

	for _, acc := range storage.State {
		stateb3[acc.Address] = *acc
	}

	paramb3 = make([]parameters, len(parameterSlice))
	copy(paramb3, parameterSlice)

	b4 := newBlock(b3.Hash)
	createBlockWithTxs(b4)
	finalizeBlock(b4)
	if err := validateBlock(b4); err != nil {
		t.Errorf("Block failed: %v\n", b4)
	}

	//STARTING ROLLBACKS---------------------------------------------
	if err := validateBlockRollback(b4); err != nil {
		t.Errorf("%v\n", err)
	}
	for _, acc := range storage.State {
		tmpState[acc.Address] = *acc
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
	for _, acc := range storage.State {
		tmpState[acc.Address] = *acc
	}
	if !reflect.DeepEqual(tmpState, stateb2) || !reflect.DeepEqual(paramb2, parameterSlice) {
		t.Error("Block rollback failed.")
	}
	for k := range tmpState {
		delete(tmpState, k)
	}

	if err := validateBlockRollback(b2); err != nil {
		t.Errorf("%v\n", err)
	}
	for _, acc := range storage.State {
		tmpState[acc.Address] = *acc
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
	for _, acc := range storage.State {
		tmpState[acc.Address] = *acc
	}

	for k := range tmpState {
		delete(tmpState, k)
	}
}
