package bc

import "fmt"

type Account struct {
	Hash [32]byte
	Address [64]byte
	TxCnt, Balance uint64
}

func (acc Account) String() string {
	return fmt.Sprintf("Hash: %x, Address: %x, TxCnt: %v, Balance: %v", acc.Hash[0:8], acc.Address[0:8], acc.TxCnt, acc.Balance)
}