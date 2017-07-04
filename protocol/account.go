package protocol

import (
	"encoding/binary"
	"fmt"
)

//testing
const (
	PubA1 = "c2be9abbeaec39a066c2a09cee23bb9ab2a0b88f2880b1e785b4d317adf0dc7c"
	PubA2 = "8ce020fde838d9c443f6c93345dafe7fd74f091c4d2f30b37e2453679a257ed5"
	PrivA = "ba127fa8f802b008b9cdb58f4e44809d48f1b000cff750dda9cd6b312395c1c5"
	PubB1 = "5d7eefd58e3d2f309471928ab4bbd104e52973372c159fa652b8ca6b57ff68b8"
	PubB2 = "ab301a6a77b201c416ddc13a2d33fdf200a5302f6f687e0ea09085debaf8a1d9"
	PrivB = "7a0a9babcc97ea7991ed67ed7f800f70c5e04e99718960ad8efab2ca052f00c7"
)

const (
	//P-256
	RootPub1 = "6323cc034597195ae69bcfb628ecdffa5989c7503154c566bab4a87f3e9910ac"
	RootPub2 = "f6115b77a15852764c609c6a5c1739e698ebc6e49bf14617c561b9110039cec7"
	RootPriv = "277ed539f56122c25a6fc115d07d632b47e71416c9aebf1beb54ee704f11842c"
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
