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
	Nonce uint64
	Amount uint32
	From, To [32]byte
}

func constrTx(txCnt uint64, amount uint32, from, to [32]byte, key *ecdsa.PrivateKey) (tx Transaction, err error) {

	//avoid sending money to its own acc, doesn't make sense with account-based money
	if reflect.DeepEqual(from,to) {
		return
	}

	//encoding nonce,from,to,amount into byte array
	//serialized := encodeTxContent(nonce, amount, from.Id, to.Id)
	sigHash := serializeHashContent(TxInfo{txCnt, amount, from, to})

	r,s, err := ecdsa.Sign(rand.Reader, key, sigHash[:])

	//this will later be DER-encoded (also ECDSA pubkey)
	copy(tx.Sig[:32],r.Bytes())
	copy(tx.Sig[32:],s.Bytes())


	tx.Info.Nonce = State[from].TxCnt
	tx.Info.From = from
	tx.Info.To = to
	tx.Info.Amount = amount

	return
}

func (tx Transaction) VerifyTx() bool {
	pub1,pub2 := new(big.Int), new(big.Int)
	r,s := new(big.Int), new(big.Int)

	//this indirection is somehow needed
	fromPubKey := State[tx.Info.From].Address

	pub1.SetBytes(fromPubKey[:32])
	pub2.SetBytes(fromPubKey[32:])

	pubKey := ecdsa.PublicKey{elliptic.P256(), pub1, pub2}
	r.SetBytes(tx.Sig[:32])
	s.SetBytes(tx.Sig[32:])

	sigHash := serializeHashContent(tx.Info)

	correct := ecdsa.Verify(&pubKey,sigHash[:],r,s)
	return correct
}