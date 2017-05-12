package bc

import (
	"math/big"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
)

//just for test cases
const (
	//P-256
	rootPub1 = "6323cc034597195ae69bcfb628ecdffa5989c7503154c566bab4a87f3e9910ac"
	rootPub2 = "f6115b77a15852764c609c6a5c1739e698ebc6e49bf14617c561b9110039cec7"
	rootPriv = "277ed539f56122c25a6fc115d07d632b47e71416c9aebf1beb54ee704f11842c"
)

type accTx struct {
	Sig [64]byte
	PubKey [64]byte
}

func constrAccTx() (tx accTx, err error) {

	_rootPub1,_ := new(big.Int).SetString(rootPub1,16)
	_rootPub2,_ := new(big.Int).SetString(rootPub2,16)
	_rootPriv,_ := new(big.Int).SetString(rootPriv,16)
	rootPubKey := ecdsa.PublicKey{
		elliptic.P256(),
		_rootPub1,
		_rootPub2,
	}
	rootPrivKey := ecdsa.PrivateKey{
		rootPubKey,
		_rootPriv,
	}

	var pubKey [64]byte
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	copy(tx.PubKey[:32],priv.PublicKey.X.Bytes())
	copy(tx.PubKey[32:],priv.PublicKey.Y.Bytes())

	r,s, err := ecdsa.Sign(rand.Reader, &rootPrivKey, pubKey[:])
	copy(tx.Sig[:32],r.Bytes())
	copy(tx.Sig[32:],s.Bytes())

	return
}

func (tx *accTx) verify() bool {

	//account creation can only be done with a valid priv/pub key which is hard-coded
	pub1,_ := new(big.Int).SetString(rootPub1,16)
	pub2,_ := new(big.Int).SetString(rootPub2,16)

	r,s := new(big.Int), new(big.Int)

	r.SetBytes(tx.Sig[:32])
	s.SetBytes(tx.Sig[32:])

	pubKey := ecdsa.PublicKey{elliptic.P256(), pub1, pub2}

	correct := ecdsa.Verify(&pubKey,tx.PubKey[:],r,s)

	return correct
}

func (tx accTx) String() string {
	return fmt.Sprintf(
		"\nSig: %x\n" +
		"PubKey: %x\n",
		tx.Sig[0:4],
		tx.PubKey[0:4],
	)
}