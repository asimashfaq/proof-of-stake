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
	DIFF_INTERVAL_ID = 2
	FEE_MINIMUM_ID = 3
	BLOCK_INTERVAL_ID = 4
	BLOCK_REWARD_ID = 5

	MIN_BLOCK_SIZE = 1000 //1KB
	MAX_BLOCK_SIZE = 100000000 //100MB

	MIN_DIFF_INTERVAL = 1 //10min for 1min interval
	MAX_DIFF_INTERVAL = 9223372036854775807

	MIN_FEE_MINIMUM = 0
	MAX_FEE_MINIMUM = 9223372036854775807

	MIN_BLOCK_INTERVAL = 30 //30 seconds
	MAX_BLOCK_INTERVAL = 86400 //24 hours

	MIN_BLOCK_REWARD = 0
	MAX_BLOCK_REWARD = 1152921504606846976 //2^60
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
			return boundsChecking(tx.Id,tx.Payload)
		}
	}

	return false
}

//returns if id is in the list of possible ids and rational value for payload parameter
func boundsChecking(id uint8, payload uint64) bool {

	switch id {
	case BLOCK_SIZE_ID:
		if payload >= MIN_BLOCK_SIZE && payload <= MAX_BLOCK_SIZE {
			return true
		}
		return false
	case DIFF_INTERVAL_ID:
		if payload >= MIN_DIFF_INTERVAL && payload <= MAX_DIFF_INTERVAL {
			return true
		}
		return false
	case FEE_MINIMUM_ID:
		if payload >= MIN_FEE_MINIMUM && payload <= MAX_FEE_MINIMUM {
			return true
		}
		return false
	case BLOCK_INTERVAL_ID:
		if payload >= MIN_BLOCK_INTERVAL && payload <= MAX_BLOCK_INTERVAL {
			return true
		}
		return false
	case BLOCK_REWARD_ID:
		if payload >= MIN_BLOCK_REWARD && payload <= MAX_BLOCK_REWARD {
			return true
		}
		return false
	default:
		return false
	}
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

func (param parameters) String() string {
	return fmt.Sprintf(
	"\n" +
		"Block Hash: %x\n" +
		"Block size: %v\n" +
		"Difficulty interval: %v\n" +
		"Fee minimum: %v\n" +
		"Block interval: %v\n" +
		"Block reward: %v\n",
		param.blockHash[0:8],
		param.block_size,
		param.diff_interval,
		param.fee_minimum,
		param.block_interval,
		param.block_reward,
	)
}