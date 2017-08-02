package protocol

import (
	"encoding/binary"
	"fmt"
)

const (
	ACC_SIZE = 76
)

type Account struct {
	Address [64]byte
	Balance uint64
	TxCnt   uint32
}

func (acc *Account) Hash() (hash [32]byte) {

	if acc == nil {
		return [32]byte{}
	}
	return serializeHashContent(acc.Address)
}

func (acc *Account) Encode() (encodedAcc []byte) {

	if acc == nil {
		return nil
	}

	encodedAcc = make([]byte, ACC_SIZE)

	var balanceBuf [8]byte
	var txCntBuf [4]byte

	binary.BigEndian.PutUint64(balanceBuf[:], acc.Balance)
	binary.BigEndian.PutUint32(txCntBuf[:], acc.TxCnt)
	copy(encodedAcc[0:64], acc.Address[:])
	copy(encodedAcc[64:72], balanceBuf[:])
	copy(encodedAcc[72:76], txCntBuf[:])

	return encodedAcc
}

func (*Account) Decode(encodedAcc []byte) (acc *Account) {

	if len(encodedAcc) != ACC_SIZE {
		return nil
	}

	acc = new(Account)
	copy(acc.Address[:], encodedAcc[0:64])
	acc.Balance = binary.BigEndian.Uint64(encodedAcc[64:72])
	acc.TxCnt = binary.BigEndian.Uint32(encodedAcc[72:76])

	return acc
}

func (acc Account) String() string {
	addressHash := serializeHashContent(acc.Address)
	return fmt.Sprintf("Hash: %x, Address: %x, TxCnt: %v, Balance: %v", addressHash[0:12], acc.Address[0:8], acc.TxCnt, acc.Balance)
}
