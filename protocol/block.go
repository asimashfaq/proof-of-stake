package protocol

import (
	"bytes"
	"encoding/binary"
	"golang.org/x/crypto/sha3"
	"fmt"
)

const (
	HASH_LEN         = 32
	PROOF_SIZE       = 9
	BLOCKHEADER_SIZE = 151
)

type transaction interface {
	verify() bool
}

type Block struct {
	Header      byte
	Hash        [32]byte
	PrevHash    [32]byte
	Nonce       [PROOF_SIZE]byte //72-bit, enough even if the network gets really large
	Timestamp   int64
	MerkleRoot  [32]byte
	Beneficiary [32]byte
	NrFundsTx   uint16
	NrAccTx     uint16
	NrConfigTx  uint8
	//this field will not be exported, this is just to avoid race conditions for the global state
	stateCopy    map[[32]byte]*Account
	FundsTxData  [][32]byte
	AccTxData    [][32]byte
	ConfigTxData [][32]byte
}

func hashBlock(b *Block) (hash [32]byte) {

	var buf bytes.Buffer

	blockToHash := struct {
		prevHash    [32]byte
		header      uint8
		timestamp   int64
		merkleRoot  [32]byte
		beneficiary [32]byte
	}{
		b.PrevHash,
		b.Header,
		b.Timestamp,
		b.MerkleRoot,
		b.Beneficiary,
	}

	binary.Write(&buf, binary.BigEndian, blockToHash)
	return sha3.Sum256(buf.Bytes())
}

func encodeBlock(b *Block) (encodedBlock []byte) {

	if b == nil {
		return nil
	}

	//making byte array of all non-byte data
	var timeStamp [8]byte
	var nrFundsTx, nrAccTx [2]byte

	binary.BigEndian.PutUint64(timeStamp[:], uint64(b.Timestamp))
	binary.BigEndian.PutUint16(nrFundsTx[:], b.NrFundsTx)
	binary.BigEndian.PutUint16(nrAccTx[:], b.NrAccTx)

	//reserve space
	encodedBlock = make([]byte,
		BLOCKHEADER_SIZE+
			int(b.NrAccTx)*HASH_LEN+
			int(b.NrFundsTx)*HASH_LEN+
			int(b.NrConfigTx)*HASH_LEN)

	encodedBlock[0] = b.Header

	copy(encodedBlock[1:33], b.Hash[:])
	copy(encodedBlock[33:65], b.PrevHash[:])
	copy(encodedBlock[65:74], b.Nonce[:])
	copy(encodedBlock[74:82], timeStamp[:])
	copy(encodedBlock[82:114], b.MerkleRoot[:])
	copy(encodedBlock[114:146], b.Beneficiary[:])
	copy(encodedBlock[146:148], nrFundsTx[:])
	copy(encodedBlock[148:150], nrAccTx[:])
	encodedBlock[150] = byte(b.NrConfigTx)

	index := BLOCKHEADER_SIZE

	for _, txHash := range b.FundsTxData {
		copy(encodedBlock[index:index+HASH_LEN], txHash[:])
		index += HASH_LEN
	}

	for _, txHash := range b.AccTxData {
		copy(encodedBlock[index:index+HASH_LEN], txHash[:])
		index += HASH_LEN
	}

	for _, txHash := range b.ConfigTxData {
		copy(encodedBlock[index:index+HASH_LEN], txHash[:])
		index += HASH_LEN
	}

	return encodedBlock
}

func decodeBlock(encodedBlock []byte) (b *Block) {

	b = new(Block)

	//time.Now().Unix() return int64, but binary.BigEndian only offers uint64
	var timeStampTmp uint64
	var timeStamp int64
	var nrFundsTx, nrAccTx uint16

	if len(encodedBlock) < BLOCKHEADER_SIZE {
		return nil
	}

	timeStampTmp = binary.BigEndian.Uint64(encodedBlock[74:82])
	nrFundsTx = binary.BigEndian.Uint16(encodedBlock[146:148])
	nrAccTx = binary.BigEndian.Uint16(encodedBlock[148:150])
	timeStamp = int64(timeStampTmp)

	b.Header = encodedBlock[0]
	copy(b.Hash[:], encodedBlock[1:33])
	copy(b.PrevHash[:], encodedBlock[33:65])
	copy(b.Nonce[:], encodedBlock[65:74])
	b.Timestamp = timeStamp
	copy(b.MerkleRoot[:], encodedBlock[82:114])
	copy(b.Beneficiary[:], encodedBlock[114:146])
	b.NrFundsTx = nrFundsTx
	b.NrAccTx = nrAccTx
	b.NrConfigTx = uint8(encodedBlock[150])

	index := BLOCKHEADER_SIZE

	var hash [32]byte
	for cnt := 0; cnt < int(nrFundsTx); cnt++ {
		copy(hash[:], encodedBlock[index:index+HASH_LEN])
		b.FundsTxData = append(b.FundsTxData, hash)
		index += HASH_LEN
	}

	for cnt := 0; cnt < int(nrAccTx); cnt++ {
		copy(hash[:], encodedBlock[index:index+HASH_LEN])
		b.AccTxData = append(b.AccTxData, hash)
		index += HASH_LEN
	}

	for cnt := 0; cnt < int(b.NrConfigTx); cnt++ {
		copy(hash[:], encodedBlock[index:index+HASH_LEN])
		b.ConfigTxData = append(b.ConfigTxData, hash)
		index += HASH_LEN
	}

	return b
}

func (b Block) String() string {
	return fmt.Sprintf("\nHash: %x\n"+
		"Previous Hash: %x\n"+
		"Header: %v\n"+
		"Nonce: %x\n"+
		"Timestamp: %v\n"+
		"MerkleRoot: %x\n"+
		"Beneficiary: %x\n"+
		"Amount of fundsTx: %v\n"+
		"Amount of accTx: %v\n"+
		"Amount of configTx: %v\n",
		b.Hash[0:8],
		b.PrevHash[0:8],
		b.Header,
		b.Nonce,
		b.Timestamp,
		b.MerkleRoot[0:8],
		b.Beneficiary[0:8],
		b.NrFundsTx,
		b.NrAccTx,
		b.NrConfigTx,
	)
}
