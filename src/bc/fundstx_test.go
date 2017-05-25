package bc

import (
	"testing"
	"fmt"
	"crypto/ecdsa"
	"crypto/rand"
	"bytes"
	"encoding/binary"
	"math/big"
	"crypto/elliptic"
)

func TestSigningVerification(t *testing.T) {

	var buf bytes.Buffer
	var sig [64]byte
	r2,s2 := new(big.Int), new(big.Int)
	pub1,pub2 := new(big.Int), new(big.Int)

	var header byte
	var amount uint32
	var fee uint16
	var txCnt uint32

	var amountBuf [4]byte
	var tmpTxCntBuf [4]byte
	var txCntBuf [3]byte
	var feeBuf [2]byte

	for i := 0; i <= 10000; i++ {


		header = 0x01
	amount = 10
	fee = 2
	txCnt = uint32(i)

	binary.Write(&buf, binary.BigEndian, fee)
	copy(feeBuf[:],buf.Bytes())
	buf.Reset()
	binary.Write(&buf, binary.BigEndian, amount)
	copy(amountBuf[:],buf.Bytes())
	buf.Reset()
	binary.Write(&buf, binary.BigEndian, txCnt)
	copy(tmpTxCntBuf[:],buf.Bytes())
	copy(txCntBuf[:],tmpTxCntBuf[1:])
	buf.Reset()

	txToHash := struct {
		Header byte
		Amount [4]byte
		Fee [2]byte
		TxCnt [3]byte
		From [32]byte
		To [32]byte
	} {
		header,
		amountBuf,
		feeBuf,
		txCntBuf,
		accA.Hash,
		accB.Hash,
	}

	sigHash := serializeHashContent(txToHash)

		r,s,err := ecdsa.Sign(rand.Reader, &PrivKeyA, sigHash[:])

		if err != nil {
			fmt.Printf("%v\n", err)
		}

		copy(sig[:32],r.Bytes())
		copy(sig[32:],s.Bytes())

		r2.SetBytes(sig[:32])
		s2.SetBytes(sig[32:])

		pub1.SetBytes(accA.Address[:32])
		pub2.SetBytes(accA.Address[32:])

		pubKey := ecdsa.PublicKey{elliptic.P256(), pub1, pub2}
		if !ecdsa.Verify(&pubKey,sigHash[:],r2,s2) {
			fmt.Errorf("r: %x vs. s: %x\n", r,s)
			fmt.Errorf("r2: %x vs. s2: %x\n", r2,s2)
		}
	}


}

func TestFundsTx(t *testing.T) {

	for i := 0; i < 1000; i++ {
		tx, _ := ConstrFundsTx(0x01, 2, 3, uint32(1), accA.Hash, accB.Hash, &PrivKeyA)
		if tx.verify() == false {
			t.Errorf("Tx could not be verified: \n%v", tx)
		}
	}
}

