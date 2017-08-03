package storage

import (
	"bytes"
	"encoding/binary"
	"golang.org/x/crypto/sha3"
	"github.com/lisgie/bazo_miner/protocol"
)

//Needed by miner and p2p package
func GetAccountFromHash(hash [32]byte) *protocol.Account { return State[hash] }

//Serializes the input in big endian and returns the sha3 hash function applied on ths input
func serializeHashContent(data interface{}) (hash [32]byte) {

	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, data)
	return sha3.Sum256(buf.Bytes())
}
