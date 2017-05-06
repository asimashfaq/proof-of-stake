package bc

import (
	"bytes"
	"encoding/binary"
	"golang.org/x/crypto/sha3"
)

func serializeHashTxContent(tx TxInfo) (hash [32]byte) {
	// Create a struct and write it.
	var buf bytes.Buffer

	binary.Write(&buf,binary.LittleEndian, tx)

	return sha3.Sum256(buf.Bytes())
}