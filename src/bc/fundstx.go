package bc

import (

	"crypto/ecdsa"
	"crypto/rand"
	"math/big"
	"crypto/elliptic"
	"reflect"
	"fmt"
)

//when we broadcast transactions we need a way to distinguish with a type

type fundsTx struct {
	Header byte
	Amount [4]byte
	TxCnt [3]byte
	From [8]byte
	To [8]byte
	Xored [24]byte
	Sig [40]byte
}

func constrFundsTx(header byte, amount [4]byte, txCnt [3]byte, from, to [32]byte, key *ecdsa.PrivateKey) (tx fundsTx, err error) {

	//avoid sending money to its own acc, doesn't make sense with account-based money
	txToHash := struct {
		Header byte
		Amount [4]byte
		TxCnt [3]byte
		From [32]byte
		To [32]byte
	} {
		header,
		amount,
		txCnt,
		from,
		to,
	}

	sigHash := serializeHashContent(txToHash)

	fmt.Printf("Constr. Hash: %x\n", sigHash)

	r,s, err := ecdsa.Sign(rand.Reader, key, sigHash[:])

	fmt.Printf("r, s: %x, %x\n", r,s)
	var sig [64]byte
	copy(sig[0:32],r.Bytes())
	copy(sig[32:],s.Bytes())

	tx.Header = header
	tx.Amount = amount
	tx.TxCnt = txCnt

	copy(tx.From[0:8],from[0:8])
	copy(tx.To[0:8],to[0:8])

	for i := 0; i < 24; i++ {
		tx.Xored[i] = from[i+8] ^ to[i+8] ^ sig[i]
	}

	copy(tx.Sig[:],sig[24:64])

	return
}

//state access should be avoided, thus the public key
func (tx *fundsTx) verify() bool {

	var sig [24]byte
	var concatSig [64]byte
	pub1,pub2 := new(big.Int), new(big.Int)
	r,s := new(big.Int), new(big.Int)

	for _,accFrom := range State[tx.From] {
		for _,accTo := range State[tx.To] {
			sig = [24]byte{}
			for cnt := 0; cnt < 24; cnt++ {
				sig[cnt] = tx.Xored[cnt] ^ accFrom.Hash[cnt+8] ^ accTo.Hash[cnt+8]
			}
			copy(concatSig[:24],sig[0:24])
			copy(concatSig[24:],tx.Sig[:])

			pub1.SetBytes(accFrom.Address[:32])
			pub2.SetBytes(accFrom.Address[32:])

			r.SetBytes(concatSig[:32])
			s.SetBytes(concatSig[32:])

			txHash := struct {
				Header byte
				Amount [4]byte
				TxCnt [3]byte
				From [32]byte
				To [32]byte
			} {
				tx.Header,
				tx.Amount,
				tx.TxCnt,
				accFrom.Hash,
				accTo.Hash,
			}
			sigHash := serializeHashContent(txHash)

			pubKey := ecdsa.PublicKey{elliptic.P256(), pub1, pub2}
			if ecdsa.Verify(&pubKey,sigHash[:],r,s) == true && !reflect.DeepEqual(accFrom,accTo) {
				return true
			}
		}
	}

	return false
}

func (tx fundsTx) String() string {
	return fmt.Sprintf(
		"Header: %x\n" +
			"Amount: %v\n" +
			"TxCnt: %v\n" +
			"From: %x\n" +
			"To: %x\n" +
			"Xored: %x\n" +
			"Sig: %x\n",
		tx.Header,
		tx.Amount,
		tx.TxCnt,
		tx.From,
		tx.To,
		tx.Xored[0:8],
		tx.Sig[0:8],
	)
}