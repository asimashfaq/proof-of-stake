package bc

import (
	"math/big"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"encoding/binary"
	"bytes"
	"log"
)

//just for test cases
const (
	//P-256
	RootPub1 = "6323cc034597195ae69bcfb628ecdffa5989c7503154c566bab4a87f3e9910ac"
	RootPub2 = "f6115b77a15852764c609c6a5c1739e698ebc6e49bf14617c561b9110039cec7"
	RootPriv = "277ed539f56122c25a6fc115d07d632b47e71416c9aebf1beb54ee704f11842c"
)

const(
	ACCTX_SIZE = 169
)

type accTx struct {
	Header byte
	Issuer [32]byte
	Fee [8]byte
	Sig [64]byte
	PubKey [64]byte
}

func ConstrAccTx(fee uint64, rootPrivKey *ecdsa.PrivateKey) (tx *accTx, err error) {

	var buf bytes.Buffer

	tx = new(accTx)

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

func EncodeAccTx(tx *accTx) (encodedTx []byte) {
	encodedTx = make([]byte,ACCTX_SIZE)
	encodedTx[0] = tx.Header
	copy(encodedTx[1:33], tx.Issuer[:])
	copy(encodedTx[33:41], tx.Fee[:])
	copy(encodedTx[41:105], tx.Sig[:])
	copy(encodedTx[105:169], tx.PubKey[:])

	return encodedTx
}

func DecodeAccTx(encodedTx []byte) (tx *accTx) {

	if len(encodedTx) < ACCTX_SIZE {
		log.Printf("DecodeAccTx, received buffer is too short: %v\n", len(encodedTx))
		return nil
	}

	tx = new(accTx)
	tx.Header = encodedTx[0]
	copy(tx.Issuer[:],encodedTx[1:33])
	copy(tx.Fee[:],encodedTx[33:41])
	copy(tx.Sig[:],encodedTx[41:105])
	copy(tx.PubKey[:],encodedTx[105:169])

	return tx
}

func (tx accTx) String() string {
	return fmt.Sprintf(
		"\n" +
			"Issuer: %x\n" +
			"Fee: %v\n" +
			"Sig: %x\n" +
			"PubKey: %x\n\n",
		tx.Issuer[0:8],
		tx.Fee,
		tx.Sig[0:8],
		tx.PubKey[0:8],
	)
}