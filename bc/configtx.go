package bc

import (
	"log"
	"fmt"
	"crypto/ecdsa"
	"crypto/elliptic"
	"math/big"
	"bytes"
	"encoding/binary"
	"crypto/rand"
)

const(
	CONFIGTX_SIZE = 82
	BLOCK_SIZE_ID = 1
	DIFFICULTY_INTERVAL_ID = 2
	FEE_MINIMUM_ID = 3
)

type configTx struct {
	Header uint8
	Id uint8
	Payload uint64
	Fee uint64
	Sig [64]byte
}

func ConstrConfigTx(header uint8, id uint8, payload uint64, fee uint64, rootPrivKey *ecdsa.PrivateKey) (tx *configTx, err error) {

	tx = new(configTx)
	tx.Header = header
	tx.Id = id
	tx.Payload = payload
	tx.Fee = fee

	txHash := hashConfigTx(tx)

	r,s,err := ecdsa.Sign(rand.Reader, rootPrivKey, txHash[:])

	if err != nil {
		return nil,err
	}

	copy(tx.Sig[32-len(r.Bytes()):32],r.Bytes())
	copy(tx.Sig[64-len(s.Bytes()):],s.Bytes())

	return tx,nil
}

func (tx *configTx) verify() bool {

	//account creation can only be done with a valid priv/pub key which is hard-coded
	r,s := new(big.Int), new(big.Int)
	pub1,pub2 := new(big.Int), new(big.Int)

	r.SetBytes(tx.Sig[:32])
	s.SetBytes(tx.Sig[32:])

	for _,rootAcc := range RootKeys {
		pub1.SetBytes(rootAcc.Address[:32])
		pub2.SetBytes(rootAcc.Address[32:])

		pubKey := ecdsa.PublicKey{elliptic.P256(), pub1, pub2}
		txHash := hashConfigTx(tx)
		if ecdsa.Verify(&pubKey,txHash[:],r,s) == true {
			return true
		}
	}

	return false
}

func hashConfigTx(tx *configTx) (hash [32]byte) {

	if tx == nil {
		return [32]byte{}
	}

	txHash := struct {
		Header uint8
		Id uint8
		Payload uint64
		Fee uint64
	} {
		tx.Header,
		tx.Id,
		tx.Payload,
		tx.Fee,
	}
	return serializeHashContent(txHash)
}

func EncodeConfigTx(tx *configTx) (encodedTx []byte) {

	if tx == nil {
		return nil
	}

	var buf bytes.Buffer
	var payloadBuf [8]byte
	var feeBuf [8]byte

	binary.Write(&buf, binary.BigEndian, tx.Payload)
	copy(payloadBuf[:],buf.Bytes())
	buf.Reset()
	binary.Write(&buf, binary.BigEndian, tx.Fee)
	copy(feeBuf[:],buf.Bytes())
	buf.Reset()

	encodedTx = make([]byte,CONFIGTX_SIZE)
	encodedTx[0] = tx.Header
	encodedTx[1] = tx.Id
	copy(encodedTx[2:10],payloadBuf[:])
	copy(encodedTx[10:18],feeBuf[:])
	copy(encodedTx[18:82],tx.Sig[:])

	return encodedTx
}

func DecodeConfigTx(encodedTx []byte) (tx *configTx) {

	if len(encodedTx) < CONFIGTX_SIZE {
		log.Printf("DecodeConfigTx, received buffer is too short: %v\n", len(encodedTx))
		return nil
	}

	tx = new(configTx)
	tx.Header = encodedTx[0]
	tx.Id = encodedTx[1]
	tx.Payload = binary.BigEndian.Uint64(encodedTx[2:10])
	tx.Fee = binary.BigEndian.Uint64(encodedTx[10:18])
	copy(tx.Sig[:],encodedTx[18:82])

	return tx
}

func (tx configTx) String() string {
	return fmt.Sprintf(
		"\n" +
			"Id: %v\n" +
			"Payload: %v\n" +
			"Fee: %v\n\n",
		tx.Id,
		tx.Payload,
		tx.Fee,
	)
}
