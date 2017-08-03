package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"golang.org/x/crypto/sha3"
)

const (
	HASH_LEN         = 32
	BLOCKHEADER_SIZE = 150
)

type transaction interface {
	verify() bool
}

type Block struct {
	Header      byte
	Hash        [32]byte
	PrevHash    [32]byte
	Nonce       [8]byte //72-bit, enough even if the network gets really large
	Timestamp   int64
	MerkleRoot  [32]byte
	Beneficiary [32]byte
	NrFundsTx   uint16
	NrAccTx     uint16
	NrConfigTx  uint8
	//this field will not be exported, this is just to avoid race conditions for the global state
	StateCopy    map[[32]byte]*Account //won't be serialized, just keeping track of local state changes
	FundsTxData  [][32]byte
	AccTxData    [][32]byte
	ConfigTxData [][32]byte
}

//Just Hash() conflicts with struct field
func (b *Block) HashBlock() (hash [32]byte) {

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

func (b *Block) GetSize() (size uint64) {
	return uint64(BLOCKHEADER_SIZE+
		int(b.NrAccTx)*HASH_LEN+
		int(b.NrFundsTx)*HASH_LEN+
		int(b.NrConfigTx)*HASH_LEN)
}

func (b *Block) Encode() (encodedBlock []byte) {

	if b == nil {
		return nil
	}

	//Making byte array of all non-byte data
	var timeStamp [8]byte
	var nrFundsTx, nrAccTx [2]byte

	binary.BigEndian.PutUint64(timeStamp[:], uint64(b.Timestamp))
	binary.BigEndian.PutUint16(nrFundsTx[:], b.NrFundsTx)
	binary.BigEndian.PutUint16(nrAccTx[:], b.NrAccTx)

	//Allocate memory
	encodedBlock = make([]byte,b.GetSize())

	encodedBlock[0] = b.Header

	copy(encodedBlock[1:33], b.Hash[:])
	copy(encodedBlock[33:65], b.PrevHash[:])
	copy(encodedBlock[65:73], b.Nonce[:])
	copy(encodedBlock[73:81], timeStamp[:])
	copy(encodedBlock[81:113], b.MerkleRoot[:])
	copy(encodedBlock[113:145], b.Beneficiary[:])
	copy(encodedBlock[145:147], nrFundsTx[:])
	copy(encodedBlock[147:149], nrAccTx[:])
	encodedBlock[149] = byte(b.NrConfigTx)

	index := BLOCKHEADER_SIZE

	//Serialize all tx hashes
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

func (*Block) Decode(encodedBlock []byte) (b *Block) {

	b = new(Block)

	if len(encodedBlock) < BLOCKHEADER_SIZE {
		return nil
	}

	timeStampTmp := binary.BigEndian.Uint64(encodedBlock[73:81])
	nrFundsTx := binary.BigEndian.Uint16(encodedBlock[145:147])
	nrAccTx := binary.BigEndian.Uint16(encodedBlock[147:149])
	timeStamp := int64(timeStampTmp)

	b.Header = encodedBlock[0]
	copy(b.Hash[:], encodedBlock[1:33])
	copy(b.PrevHash[:], encodedBlock[33:65])
	copy(b.Nonce[:], encodedBlock[65:73])
	b.Timestamp = timeStamp
	copy(b.MerkleRoot[:], encodedBlock[81:113])
	copy(b.Beneficiary[:], encodedBlock[113:145])
	b.NrFundsTx = nrFundsTx
	b.NrAccTx = nrAccTx
	b.NrConfigTx = uint8(encodedBlock[149])

	index := BLOCKHEADER_SIZE

	//Deserialize all tx hashes
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
