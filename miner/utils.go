package miner

import (
	"bytes"
	"encoding/binary"
	"github.com/lisgie/bazo_miner/protocol"
	"github.com/lisgie/bazo_miner/storage"
	"golang.org/x/crypto/sha3"
)

//Was thinking of whether to put that into the storage package. But I think this is a more suitable place
func getAccountFromHash(hash [32]byte) *protocol.Account { return storage.State[hash] }

//Serializes the input in big endian and returns the sha3 hash function applied on ths input
func serializeHashContent(data interface{}) (hash [32]byte) {

	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, data)
	return sha3.Sum256(buf.Bytes())
}
