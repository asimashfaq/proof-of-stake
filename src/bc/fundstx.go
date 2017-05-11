package bc

import (
	"math/big"
	"reflect"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/elliptic"
)

//when we broadcast transactions we need a way to distinguish with a type
type fundsTx struct {
	Sig [64]byte
	Payload txPayload
}

type txPayload struct {
	Nonce uint64
	Amount uint32
	From, To [32]byte
}

func constrFundsTx(txCnt uint64, amount uint32, from, to [32]byte, key *ecdsa.PrivateKey) (tx fundsTx, err error) {

	//avoid sending money to its own acc, doesn't make sense with account-based money
	if reflect.DeepEqual(from,to) {
		return
	}

	//encoding nonce,from,to,amount into byte array
	//serialized := encodeTxContent(nonce, amount, from.Id, to.Id)
	sigHash := serializeHashContent(txPayload{txCnt, amount, from, to})

	r,s, err := ecdsa.Sign(rand.Reader, key, sigHash[:])

	//this will later be DER-encoded (also ECDSA pubkey)
	copy(tx.Sig[:32],r.Bytes())
	copy(tx.Sig[32:],s.Bytes())


	tx.Payload.Nonce = txCnt
	tx.Payload.From = from
	tx.Payload.To = to
	tx.Payload.Amount = amount

	return
}

//state access should be avoided, thus the public key
func (tx fundsTx) verify() bool {
	pub1,pub2 := new(big.Int), new(big.Int)
	r,s := new(big.Int), new(big.Int)

	senderAddress := State[tx.Payload.From].Address

	pub1.SetBytes(senderAddress[:32])
	pub2.SetBytes(senderAddress[32:])

	pubKey := ecdsa.PublicKey{elliptic.P256(), pub1, pub2}
	r.SetBytes(tx.Sig[:32])
	s.SetBytes(tx.Sig[32:])

	sigHash := serializeHashContent(tx.Payload)

	correct := ecdsa.Verify(&pubKey,sigHash[:],r,s)
	return correct
}