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

func ConstrAccTx(rootPrivKey *ecdsa.PrivateKey) (tx accTx, err error) {

	//fixed fee for now
	var buf bytes.Buffer
	var fee uint16

	//fee will be discarded later
	fee = 5

	binary.Write(&buf,binary.BigEndian,fee)
	copy(tx.Fee[:],buf.Bytes())

	//var pubKey [64]byte
	newAccAddress, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	newAccPub1,newAccPub2 := newAccAddress.PublicKey.X.Bytes(),newAccAddress.PublicKey.Y.Bytes()
	copy(tx.PubKey[32-len(newAccPub1):32],newAccPub1)
	copy(tx.PubKey[64-len(newAccPub2):],newAccPub2)

	r,s, err := ecdsa.Sign(rand.Reader, rootPrivKey, tx.PubKey[:])

	var rootPublicKey [64]byte
	rootPubKey1,rootPubKey2 := rootPrivKey.PublicKey.X.Bytes(),rootPrivKey.PublicKey.Y.Bytes()
	copy(rootPublicKey[32-len(rootPubKey1):32],rootPubKey1)
	copy(rootPublicKey[64-len(rootPubKey2):],rootPubKey2)

	issuer := serializeHashContent(rootPublicKey)
	copy(tx.Issuer[:], issuer[:])

	copy(tx.Sig[32-len(r.Bytes()):32],r.Bytes())
	copy(tx.Sig[64-len(s.Bytes()):],s.Bytes())

	return
}

func (tx *accTx) verify() bool {

	//account creation can only be done with a valid priv/pub key which is hard-coded
	r,s := new(big.Int), new(big.Int)
	pub1,pub2 := new(big.Int), new(big.Int)

	r.SetBytes(tx.Sig[:32])
	s.SetBytes(tx.Sig[32:])

	for _,rootAcc := range RootKeys {
		pub1.SetBytes(rootAcc.Address[:32])
		pub2.SetBytes(rootAcc.Address[32:])

		pubKey := ecdsa.PublicKey{elliptic.P256(), pub1, pub2}
		if ecdsa.Verify(&pubKey,tx.PubKey[:],r,s) == true {
			return true
		}
	}

	return false
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