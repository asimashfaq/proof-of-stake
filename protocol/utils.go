package protocol

import (
	"bytes"
	"encoding/binary"
	"golang.org/x/crypto/sha3"
)

func serializeHashContent(data interface{}) (hash [32]byte) {

	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, data)
	return sha3.Sum256(buf.Bytes())
}

func ignoreMSB(input *uint64) uint64 {
	return (*input) & 0x7FFFFFFFFFFFFFFF
}