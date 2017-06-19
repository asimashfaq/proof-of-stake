package protocol

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
)

const (
	FUNDSTX_SIZE = 101
)

//when we broadcast transactions we need a way to distinguish with a type

type FundsTx struct {
	Header   byte
	Amount   uint64
	Fee      uint64
	TxCnt    uint32
	From     [8]byte
	FromHash [32]byte
	To       [8]byte
	ToHash   [32]byte
	Xored    [24]byte
	Sig      [40]byte
}

func ConstrFundsTx(header byte, amount uint64, fee uint64, txCnt uint32, from, to [32]byte, key *ecdsa.PrivateKey) (tx *FundsTx, err error) {

	tx = new(FundsTx)

	tx.FromHash = from
	tx.ToHash = to
	tx.Header = header
	tx.Amount = amount
	tx.Fee = fee
	tx.TxCnt = txCnt

	copy(tx.From[0:8], from[0:8])
	copy(tx.To[0:8], to[0:8])

	txHash := tx.Hash()

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

func (tx *FundsTx) Hash() (hash [32]byte) {

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
		tx.FromHash,
		tx.ToHash,
	}
	return serializeHashContent(txHash)
}

//when we serialize the struct with binary.Write, unexported field get serialized as well, undesired
//behavior. Therefore, writing own encoder/decoder
func (tx *FundsTx) Encode() (encodedTx []byte) {

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

func (*FundsTx) Decode(encodedTx []byte) (tx *FundsTx) {

	if len(encodedTx) < FUNDSTX_SIZE {
		log.Printf("DecodeFundsTxTx, received buffer is too short: %v\n", len(encodedTx))
		return nil
	}

	tx = new(FundsTx)
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

func (tx *FundsTx) TxFee() uint64 {
	return tx.Fee
}

func (tx FundsTx) String() string {
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
		tx.FromHash[0:12],
		tx.To,
		tx.ToHash[0:12],
		tx.Xored[0:8],
		tx.Sig[0:8],
	)
}
