package bc

import (
	"math/big"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"encoding/binary"
	"bytes"
)

//just for test cases
const (
	//P-256
	RootPub1 = "6323cc034597195ae69bcfb628ecdffa5989c7503154c566bab4a87f3e9910ac"
	RootPub2 = "f6115b77a15852764c609c6a5c1739e698ebc6e49bf14617c561b9110039cec7"
	RootPriv = "277ed539f56122c25a6fc115d07d632b47e71416c9aebf1beb54ee704f11842c"
)

type accTx struct {
	Issuer [32]byte
	Fee [2]byte
	Sig [64]byte
	PubKey [64]byte
}

func constrAccTx() (tx accTx, err error) {

	//fixed fee for now
	var buf bytes.Buffer
	var fee uint16
	fee = 5

	binary.Write(&buf,binary.BigEndian,fee)
	copy(tx.Fee[:],buf.Bytes())

	_rootPub1,_ := new(big.Int).SetString(RootPub1,16)
	_rootPub2,_ := new(big.Int).SetString(RootPub2,16)
	_rootPriv,_ := new(big.Int).SetString(RootPriv,16)
	rootPubKey := ecdsa.PublicKey{
		elliptic.P256(),
		_rootPub1,
		_rootPub2,
	}
	rootPrivKey := ecdsa.PrivateKey{
		rootPubKey,
		_rootPriv,
	}

	//var pubKey [64]byte
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	copy(tx.PubKey[:32],priv.PublicKey.X.Bytes())
	copy(tx.PubKey[32:],priv.PublicKey.Y.Bytes())

	r,s, err := ecdsa.Sign(rand.Reader, &rootPrivKey, tx.PubKey[:])


	var rootPublicKey [64]byte
	copy(rootPublicKey[:32],_rootPub1.Bytes())
	copy(rootPublicKey[32:],_rootPub2.Bytes())
	issuerHash := serializeHashContent(rootPublicKey)

	copy(tx.Issuer[:], issuerHash[:])

	copy(tx.Sig[:32],r.Bytes())
	copy(tx.Sig[32:],s.Bytes())

	return
}

func (tx *accTx) verify() bool {

	//account creation can only be done with a valid priv/pub key which is hard-coded
	pub1,_ := new(big.Int).SetString(RootPub1,16)
	pub2,_ := new(big.Int).SetString(RootPub2,16)

	r,s := new(big.Int), new(big.Int)

	r.SetBytes(tx.Sig[:32])
	s.SetBytes(tx.Sig[32:])

	pubKey := ecdsa.PublicKey{elliptic.P256(), pub1, pub2}

	correct := ecdsa.Verify(&pubKey,tx.PubKey[:],r,s)

	return correct
}

func (tx accTx) String() string {
	return fmt.Sprintf(
		"\n" +
			"Issuer: %x\n" +
			"Fee: %x\n" +
			"Sig: %x\n" +
			"PubKey: %x\n",
		tx.Issuer[0:8],
		tx.Fee,
		tx.Sig[0:8],
		tx.PubKey[0:8],
	)
}