package bc

import (
	"hash"
	"math/big"
	"reflect"
	"crypto/ecdsa"
	//"golang.org/x/crypto/sha3"
	"bytes"
	"crypto/rand"
	"fmt"
	"encoding/binary"
	"golang.org/x/crypto/sha3"
)

type Transaction struct {
	Hash [32]byte
	R,S   *big.Int //ecdsa sig pair
	From, To Account
	Amount int64
}

type HashBuildingBlocks struct {
	Nonce, Amount int64
	From, To [64]byte
}


func ConstrTransact(nonce, amount int64, from, to Account, key *ecdsa.PrivateKey) (tx Transaction, err error) {

	//meaningful error msg follows
	if amount > from.Balance {
		return
	}

	if reflect.DeepEqual(from,to) {
		return
	}

	if nonce != from.Nonce {
		return
	}

	//encoding nonce,from,to,amount into byte array
	serialized := encodeTransactContent(nonce, amount, from.Id, to.Id)
	tx.R, tx.S, err = ecdsa.Sign(rand.Reader, key, serialized)

	tx.From = from
	tx.To = to
	tx.Amount = amount
	tx.Hash = sha3.Sum256(serialized)

	return
}

func encodeTransactContent(nonce, amount int64, from, to [64]byte) (enc []byte) {

	// Create a struct and write it.
	var buf bytes.Buffer

	hash := HashBuildingBlocks{nonce, amount, from, to}
	binary.Write(&buf,binary.LittleEndian, hash)

	return buf.Bytes()
}