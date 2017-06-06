package bc

import (
	"fmt"
	"encoding/binary"
)

const (
	ACC_SIZE = 108
)

type Account struct {
	Hash [32]byte
	Address [64]byte
	Balance uint64
	TxCnt uint32
}

func EncodeAcc(acc *Account) (encodedAcc []byte) {

	encodedAcc = make([]byte, ACC_SIZE)

	var balanceBuf [8]byte
	var txCntBuf [4]byte

	binary.BigEndian.PutUint64(balanceBuf[:],acc.Balance)
	binary.BigEndian.PutUint32(txCntBuf[:],acc.TxCnt)
	copy(encodedAcc[0:32],acc.Hash[:])
	copy(encodedAcc[32:96],acc.Address[:])
	copy(encodedAcc[96:104],balanceBuf[:])
	copy(encodedAcc[104:108],txCntBuf[:])

	return encodedAcc
}

func DecodeAcc(encodedAcc []byte) (acc *Account) {

	acc = new(Account)
	copy(acc.Hash[:],encodedAcc[0:32])
	copy(acc.Address[:],encodedAcc[32:96])
	acc.Balance = binary.BigEndian.Uint64(encodedAcc[96:104])
	acc.TxCnt = binary.BigEndian.Uint32(encodedAcc[104:108])

	return acc
}

func (acc Account) String() string {
	return fmt.Sprintf("Hash: %x, Address: %x, TxCnt: %v, Balance: %v", acc.Hash[0:8], acc.Address[0:8], acc.TxCnt, acc.Balance)
}
