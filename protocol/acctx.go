package protocol

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	"math/big"
)

//just for test cases
const (
	//P-256
	RootPub1 = "6323cc034597195ae69bcfb628ecdffa5989c7503154c566bab4a87f3e9910ac"
	RootPub2 = "f6115b77a15852764c609c6a5c1739e698ebc6e49bf14617c561b9110039cec7"
	RootPriv = "277ed539f56122c25a6fc115d07d632b47e71416c9aebf1beb54ee704f11842c"
)

const (
	ACCTX_SIZE = 169
)

type AccTx struct {
	Header byte
	Issuer [32]byte
	Fee    uint64
	PubKey [64]byte
	Sig    [64]byte
}

func ConstrAccTx(fee uint64, rootPrivKey *ecdsa.PrivateKey) (tx *AccTx, err error) {

	tx = new(AccTx)
	tx.Fee = fee

	newAccAddress, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	newAccPub1, newAccPub2 := newAccAddress.PublicKey.X.Bytes(), newAccAddress.PublicKey.Y.Bytes()
	copy(tx.PubKey[32-len(newAccPub1):32], newAccPub1)
	copy(tx.PubKey[64-len(newAccPub2):], newAccPub2)

	var rootPublicKey [64]byte
	rootPubKey1, rootPubKey2 := rootPrivKey.PublicKey.X.Bytes(), rootPrivKey.PublicKey.Y.Bytes()
	copy(rootPublicKey[32-len(rootPubKey1):32], rootPubKey1)
	copy(rootPublicKey[64-len(rootPubKey2):], rootPubKey2)

	issuer := serializeHashContent(rootPublicKey)
	copy(tx.Issuer[:], issuer[:])

	txHash := tx.Hash()

	r, s, err := ecdsa.Sign(rand.Reader, rootPrivKey, txHash[:])
	if err != nil {
		return nil, err
	}

	copy(tx.Sig[32-len(r.Bytes()):32], r.Bytes())
	copy(tx.Sig[64-len(s.Bytes()):], s.Bytes())

	return tx, nil
}

func (tx *AccTx) Verify() bool {

	//account creation can only be done with a valid priv/pub key which is hard-coded
	r, s := new(big.Int), new(big.Int)
	pub1, pub2 := new(big.Int), new(big.Int)

	r.SetBytes(tx.Sig[:32])
	s.SetBytes(tx.Sig[32:])

	for _, rootAcc := range RootKeys {
		pub1.SetBytes(rootAcc.Address[:32])
		pub2.SetBytes(rootAcc.Address[32:])

		pubKey := ecdsa.PublicKey{elliptic.P256(), pub1, pub2}
		txHash := tx.Hash()
		if ecdsa.Verify(&pubKey, txHash[:], r, s) == true {
			return true
		}
	}

	return false
}

func (tx *AccTx) Hash() (hash [32]byte) {

	if tx == nil {
		return [32]byte{}
	}

	txHash := struct {
		Header byte
		Issuer [32]byte
		Fee    uint64
		PubKey [64]byte
	}{
		tx.Header,
		tx.Issuer,
		tx.Fee,
		tx.PubKey,
	}
	return serializeHashContent(txHash)
}

func (tx *AccTx) Encode() (encodedTx []byte) {

	if tx == nil {
		return nil
	}

	var buf bytes.Buffer
	var feeBuf [8]byte

	binary.Write(&buf, binary.BigEndian, tx.Fee)
	copy(feeBuf[:], buf.Bytes())

	encodedTx = make([]byte, ACCTX_SIZE)
	encodedTx[0] = tx.Header
	copy(encodedTx[1:33], tx.Issuer[:])
	copy(encodedTx[33:41], feeBuf[:])
	copy(encodedTx[41:105], tx.PubKey[:])
	copy(encodedTx[105:169], tx.Sig[:])

	return encodedTx
}

func (*AccTx) Decode(encodedTx []byte) (tx *AccTx) {

	if len(encodedTx) < ACCTX_SIZE {
		log.Printf("DecodeAccTx, received buffer is too short: %v\n", len(encodedTx))
		return nil
	}

	tx = new(AccTx)
	tx.Header = encodedTx[0]
	copy(tx.Issuer[:], encodedTx[1:33])
	tx.Fee = binary.BigEndian.Uint64(encodedTx[33:41])
	copy(tx.PubKey[:], encodedTx[41:105])
	copy(tx.Sig[:], encodedTx[105:169])

	return tx
}

func (tx AccTx) String() string {
	return fmt.Sprintf(
		"\n"+
			"Issuer: %x\n"+
			"Fee: %v\n"+
			"PubKey: %x\n"+
			"Sig: %x\n\n",
		tx.Issuer[0:8],
		tx.Fee,
		tx.Sig[0:8],
		tx.PubKey[0:8],
	)
}
