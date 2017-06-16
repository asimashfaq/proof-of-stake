package bc

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	"math/big"
	"reflect"
)

const (
	FUNDSTX_SIZE = 101
)

//when we broadcast transactions we need a way to distinguish with a type

type fundsTx struct {
	Header   byte
	Amount   uint64
	Fee      uint64
	TxCnt    uint32
	From     [8]byte
	fromHash [32]byte
	To       [8]byte
	toHash   [32]byte
	Xored    [24]byte
	Sig      [40]byte
}

func ConstrFundsTx(header byte, amount uint64, fee uint64, txCnt uint32, from, to [32]byte, key *ecdsa.PrivateKey) (tx *fundsTx, err error) {

	tx = new(fundsTx)

	tx.fromHash = from
	tx.toHash = to
	tx.Header = header
	tx.Amount = amount
	tx.Fee = fee
	tx.TxCnt = txCnt

	copy(tx.From[0:8], from[0:8])
	copy(tx.To[0:8], to[0:8])

	txHash := hashFundsTx(tx)

	r, s, err := ecdsa.Sign(rand.Reader, key, txHash[:])

	var sig [64]byte
	copy(sig[32-len(r.Bytes()):32], r.Bytes())
	copy(sig[64-len(s.Bytes()):], s.Bytes())

	for i := 0; i < 24; i++ {
		tx.Xored[i] = from[i+8] ^ to[i+8] ^ sig[i]
	}

	copy(tx.Sig[:], sig[24:64])

	return
}

//I believe sender balance check here is a bad idea. This limits to receive and send funds within the same block
//But if receiving and sending along funds within the same block, transaction ordering is important
func (tx *fundsTx) verify() bool {

	var sig [24]byte
	var concatSig [64]byte
	pub1, pub2 := new(big.Int), new(big.Int)
	r, s := new(big.Int), new(big.Int)

	//fundstx only makes sense if amount > 0
	if tx.Amount == 0 || tx.Amount > MAX_MONEY {
		log.Printf("Invalid transaction amount %v\n", tx.Amount)
		return false
	}

	//check if accounts are present in the actual state
	for _, accFrom := range State[tx.From] {
		accFromHash := serializeHashContent(accFrom.Address)
		for _, accTo := range State[tx.To] {
			accToHash := serializeHashContent(accTo.Address)
			sig = [24]byte{}
			for cnt := 0; cnt < 24; cnt++ {
				sig[cnt] = tx.Xored[cnt] ^ accFromHash[cnt+8] ^ accToHash[cnt+8]
			}
			copy(concatSig[:24], sig[0:24])
			copy(concatSig[24:], tx.Sig[:])

			pub1.SetBytes(accFrom.Address[:32])
			pub2.SetBytes(accFrom.Address[32:])

			r.SetBytes(concatSig[:32])
			s.SetBytes(concatSig[32:])

			tx.fromHash = accFromHash
			tx.toHash = accToHash

			txHash := hashFundsTx(tx)

			pubKey := ecdsa.PublicKey{elliptic.P256(), pub1, pub2}
			if ecdsa.Verify(&pubKey, txHash[:], r, s) == true && !reflect.DeepEqual(accFrom, accTo) {
				tx.fromHash = accFromHash
				tx.toHash = accToHash
				return true
			}
		}
	}

	return false
}

func hashFundsTx(tx *fundsTx) (hash [32]byte) {

	if tx == nil {
		//is returning nil better?
		return [32]byte{}
	}

	txHash := struct {
		Header byte
		Amount uint64
		Fee    uint64
		TxCnt  uint32
		From   [32]byte
		To     [32]byte
	}{
		tx.Header,
		tx.Amount,
		tx.Fee,
		tx.TxCnt,
		tx.fromHash,
		tx.toHash,
	}
	return serializeHashContent(txHash)
}

//when we serialize the struct with binary.Write, unexported field get serialized as well, undesired
//behavior. Therefore, writing own encoder/decoder
func EncodeFundsTx(tx *fundsTx) (encodedTx []byte) {

	if tx == nil {
		return nil
	}

	var buf bytes.Buffer
	var amountBuf [8]byte
	var feeBuf [8]byte
	var txCntBuf [4]byte

	//transfer integer values to byte arrays
	binary.Write(&buf, binary.BigEndian, tx.Amount)
	copy(amountBuf[:], buf.Bytes())
	buf.Reset()
	binary.Write(&buf, binary.BigEndian, tx.Fee)
	copy(feeBuf[:], buf.Bytes())
	buf.Reset()
	binary.Write(&buf, binary.BigEndian, tx.TxCnt)
	copy(txCntBuf[:], buf.Bytes())
	buf.Reset()

	encodedTx = make([]byte, FUNDSTX_SIZE)
	encodedTx[0] = tx.Header
	copy(encodedTx[1:9], amountBuf[:])
	copy(encodedTx[9:17], feeBuf[:])
	copy(encodedTx[17:21], txCntBuf[:])
	copy(encodedTx[21:29], tx.From[:])
	copy(encodedTx[29:37], tx.To[:])
	copy(encodedTx[37:61], tx.Xored[:])
	copy(encodedTx[61:101], tx.Sig[:])

	return encodedTx
}

func DecodeFundsTx(encodedTx []byte) (tx *fundsTx) {

	if len(encodedTx) < FUNDSTX_SIZE {
		log.Printf("DecodeFundsTxTx, received buffer is too short: %v\n", len(encodedTx))
		return nil
	}

	tx = new(fundsTx)
	tx.Header = encodedTx[0]
	tx.Amount = binary.BigEndian.Uint64(encodedTx[1:9])
	tx.Fee = binary.BigEndian.Uint64(encodedTx[9:17])
	tx.TxCnt = binary.BigEndian.Uint32(encodedTx[17:21])
	copy(tx.From[:], encodedTx[21:29])
	copy(tx.To[:], encodedTx[29:37])
	copy(tx.Xored[:], encodedTx[37:61])
	copy(tx.Sig[:], encodedTx[61:101])

	return tx
}

func (tx fundsTx) String() string {
	return fmt.Sprintf(
		"\nHeader: %x\n"+
			"Amount: %v\n"+
			"Fee: %v\n"+
			"TxCnt: %v\n"+
			"From: %x\n"+
			"From Full Hash: %x\n"+
			"To: %x\n"+
			"To Full Hash: %x\n"+
			"Xored: %x\n"+
			"Sig: %x\n\n",
		tx.Header,
		tx.Amount,
		tx.Fee,
		tx.TxCnt,
		tx.From,
		tx.fromHash[0:12],
		tx.To,
		tx.toHash[0:12],
		tx.Xored[0:8],
		tx.Sig[0:8],
	)
}
