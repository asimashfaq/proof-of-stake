package miner

import (
	"bytes"
	"encoding/binary"
	"golang.org/x/crypto/sha3"
	"github.com/lisgie/bazo_miner/protocol"
)

func getAccountFromHash(hash [32]byte) *protocol.Account {

	var fixedHash [8]byte
	copy(fixedHash[:], hash[0:8])
	for _, acc := range State[fixedHash] {
		accHash := serializeHashContent(acc.Address)
		if accHash == hash {
			return acc
		}
	}
	return nil
}

func serializeHashContent(data interface{}) (hash [32]byte) {

	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, data)
	return sha3.Sum256(buf.Bytes())
}
