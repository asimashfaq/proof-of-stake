package storage

import (
	"bytes"
	"encoding/binary"
	"github.com/lisgie/bazo_miner/protocol"
	"golang.org/x/crypto/sha3"
)

func GetAccountFromHash(hash [32]byte) *protocol.Account { return State[hash] }

func serializeHashContent(data interface{}) (hash [32]byte) {

	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, data)
	return sha3.Sum256(buf.Bytes())
}
