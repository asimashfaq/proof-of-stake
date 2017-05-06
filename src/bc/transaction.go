package bc

import (
	"math/big"
	"reflect"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/elliptic"
)

type Transaction struct {
	Sig [64]byte
	Info TxInfo
}

type TxInfo struct {
	Nonce, Amount int64
	From, To [64]byte
}


func ConstrTx(nonce, amount int64, from, to [64]byte, key *ecdsa.PrivateKey) (tx Transaction, err error) {

	//checking legal balance
	if amount <= 0 {
		return
	}

	//avoid sending money its own acc, doesn't make sense with account-based money
	if reflect.DeepEqual(from,to) {
		return
	}

	//encoding nonce,from,to,amount into byte array
	//serialized := encodeTxContent(nonce, amount, from.Id, to.Id)
	sigHash := serializeHashContent(TxInfo{nonce, amount, from, to})

	r,s, err := ecdsa.Sign(rand.Reader, key, sigHash[:])

	//this will later be DER-encoded (also ECDSA pubkey)
	copy(tx.Sig[:32],r.Bytes())
	copy(tx.Sig[32:],s.Bytes())

	tx.Info.From = from
	tx.Info.To = to
	tx.Info.Amount = amount

	return
}

func (tx *Transaction) VerifyTx() bool {
	pub1,pub2 := new(big.Int), new(big.Int)
	r,s := new(big.Int), new(big.Int)

	pub1.SetBytes(tx.Info.From[:32])
	pub2.SetBytes(tx.Info.From[32:])

	pubKey := ecdsa.PublicKey{elliptic.P256(), pub1, pub2}
	r.SetBytes(tx.Sig[:32])
	s.SetBytes(tx.Sig[32:])

	sigHash := serializeHashContent(tx.Info)

	correct := ecdsa.Verify(&pubKey,sigHash[:],r,s)

	return correct
}