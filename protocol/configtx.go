package protocol

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/binary"
	"fmt"
)

const (
	CONFIGTX_SIZE = 83

	BLOCK_SIZE_ID     = 1
	DIFF_INTERVAL_ID  = 2
	FEE_MINIMUM_ID    = 3
	BLOCK_INTERVAL_ID = 4
	BLOCK_REWARD_ID   = 5

	MIN_BLOCK_SIZE = 1000      //1KB
	MAX_BLOCK_SIZE = 100000000 //100MB

	MIN_DIFF_INTERVAL = 30 //amount in seconds
	MAX_DIFF_INTERVAL = 9223372036854775807

	MIN_FEE_MINIMUM = 0
	MAX_FEE_MINIMUM = 9223372036854775807

	MIN_BLOCK_INTERVAL = 30    //30 seconds
	MAX_BLOCK_INTERVAL = 86400 //24 hours

	MIN_BLOCK_REWARD = 0
	MAX_BLOCK_REWARD = 1152921504606846976 //2^60
)

type ConfigTx struct {
	Header  byte
	Id      uint8
	Payload uint64
	Fee     uint64
	TxCnt	uint8
	Sig     [64]byte
}

func ConstrConfigTx(header byte, id uint8, payload uint64, fee uint64, txCnt uint8, rootPrivKey *ecdsa.PrivateKey) (tx *ConfigTx, err error) {

	tx = new(ConfigTx)
	tx.Header = header
	tx.Id = id
	tx.Payload = payload
	tx.Fee = fee
	tx.TxCnt = txCnt

	txHash := tx.Hash()

	r, s, err := ecdsa.Sign(rand.Reader, rootPrivKey, txHash[:])

	if err != nil {
		return nil, err
	}

	copy(tx.Sig[32-len(r.Bytes()):32], r.Bytes())
	copy(tx.Sig[64-len(s.Bytes()):], s.Bytes())

	return tx, nil
}

func (tx *ConfigTx) Hash() (hash [32]byte) {

	if tx == nil {
		return [32]byte{}
	}

	txHash := struct {
		Header  byte
		Id      uint8
		Payload uint64
		Fee     uint64
		TxCnt	uint8
	}{
		tx.Header,
		tx.Id,
		tx.Payload,
		tx.Fee,
		tx.TxCnt,
	}
	return serializeHashContent(txHash)
}

func (tx *ConfigTx) Encode() (encodedTx []byte) {

	if tx == nil {
		return nil
	}

	var buf bytes.Buffer
	var payloadBuf [8]byte
	var feeBuf [8]byte

	binary.Write(&buf, binary.BigEndian, tx.Payload)
	copy(payloadBuf[:], buf.Bytes())
	buf.Reset()
	binary.Write(&buf, binary.BigEndian, tx.Fee)
	copy(feeBuf[:], buf.Bytes())
	buf.Reset()

	encodedTx = make([]byte, CONFIGTX_SIZE)
	encodedTx[0] = tx.Header
	encodedTx[1] = tx.Id
	copy(encodedTx[2:10], payloadBuf[:])
	copy(encodedTx[10:18], feeBuf[:])
	encodedTx[18] = byte(tx.TxCnt)
	copy(encodedTx[19:83], tx.Sig[:])

	return encodedTx
}

func (*ConfigTx) Decode(encodedTx []byte) (tx *ConfigTx) {

	if len(encodedTx) != CONFIGTX_SIZE {
		return nil
	}

	tx = new(ConfigTx)
	tx.Header = encodedTx[0]
	tx.Id = encodedTx[1]
	tx.Payload = binary.BigEndian.Uint64(encodedTx[2:10])
	tx.Fee = binary.BigEndian.Uint64(encodedTx[10:18])
	tx.TxCnt = uint8(encodedTx[18])
	copy(tx.Sig[:], encodedTx[19:83])

	return tx
}

func (tx *ConfigTx) TxFee() uint64 { return tx.Fee }
func (tx *ConfigTx) Size() uint64   { return CONFIGTX_SIZE }

func (tx ConfigTx) String() string {
	return fmt.Sprintf(
		"\n"+
			"Id: %v\n"+
			"Payload: %v\n"+
			"Fee: %v\n"+
			"TxCnt: %v\n",
		tx.Id,
		tx.Payload,
		tx.Fee,
		tx.TxCnt,
	)
}
