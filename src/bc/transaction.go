package bc

import (
	"math/big"
	"reflect"
	"crypto/ecdsa"
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"golang.org/x/crypto/sha3"
	"crypto/elliptic"
)

type Transaction struct {
	Hash [32]byte
	Sig [64]byte
	Info TxInfo
}

type TxInfo struct {
	Nonce, Amount int64
	From, To [64]byte
}


func ConstrTx(nonce, amount int64, from, to Account, key *ecdsa.PrivateKey) (tx Transaction, err error) {

	//checking legal balance
	if amount > from.Balance && amount > 0 {
		return
	}

	//avoid sending money its own acc, doesn't make sense with account-based money
	if reflect.DeepEqual(from,to) {
		return
	}

	//protecting against replay attacks
	if nonce != from.Nonce {
		return
	}

	//encoding nonce,from,to,amount into byte array
	//serialized := encodeTxContent(nonce, amount, from.Id, to.Id)
	tx.Hash = sha3.Sum256(serializeTxContent(TxInfo{nonce, amount, from.Id, to.Id}))

	r,s, err := ecdsa.Sign(rand.Reader, key, tx.Hash[:])

	//this will later be DER-encoded (also ECDSA pubkey)
	copy(tx.Sig[:32],r.Bytes())
	copy(tx.Sig[32:],s.Bytes())

	tx.Info.From = from.Id
	tx.Info.To = to.Id
	tx.Info.Amount = amount

	return
}

func serializeTxContent(tx TxInfo) (enc []byte) {
	// Create a struct and write it.
	var buf bytes.Buffer

	binary.Write(&buf,binary.LittleEndian, tx)

	return buf.Bytes()
}

func (tx *Transaction) VerifyTx() bool {
	pub1,pub2 := new(big.Int), new(big.Int)
	r,s := new(big.Int), new(big.Int)

	pub1.SetBytes(tx.Info.From[:32])
	pub2.SetBytes(tx.Info.From[32:])

	pubKey := ecdsa.PublicKey{elliptic.P256(), pub1, pub2}
	r.SetBytes(tx.Sig[:32])
	s.SetBytes(tx.Sig[32:])

	correct := ecdsa.Verify(&pubKey,tx.Hash[:],r,s)

	return correct
}