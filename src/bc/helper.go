package bc

import (
	"bytes"
	"encoding/binary"
	"golang.org/x/crypto/sha3"
	"encoding/gob"
)

func serializeHashContent(data interface{}) (hash [32]byte) {
	// Create a struct and write it.
	var buf bytes.Buffer

	binary.Write(&buf,binary.LittleEndian, data)

	return sha3.Sum256(buf.Bytes())
}

func EncodeForSend(data interface{}) []byte {

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(data)
	return buf.Bytes()
}

func DecodeForReceive(payload []byte) interface{} {

	var decoded []byte
	var buf bytes.Buffer

	dec := gob.NewDecoder(&buf)
	dec.Decode(decoded)
	return buf.Bytes()
}