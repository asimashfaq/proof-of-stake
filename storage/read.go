package storage

import (
	"github.com/boltdb/bolt"
)

type Kvtuple struct {
	//everything is stored as byte slices, it's faster to not use [32]byte, needs additional copying
	Hash []byte
	Payload []byte
}

