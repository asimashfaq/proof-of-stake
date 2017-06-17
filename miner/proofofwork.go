package miner

import (
	"golang.org/x/crypto/sha3"
	"math/big"
)

func validateProofOfWork(diff uint8, hash [32]byte) bool {
	var byteNr uint8
	for byteNr = 0; byteNr < (uint8)(diff/8); byteNr++ {
		if hash[byteNr] != 0 {
			return false
		}
	}
	if diff%8 != 0 && hash[byteNr+1] >= 1<<(8-diff%8) {
		return false
	}
	return true
}

func proofOfWork(diff uint8, partialHash [32]byte) *big.Int {

	var tmp [32]byte
	var byteNr uint8
	var abort bool
	//big int needed because int64 overflows if nonce too large
	oneIncr := big.NewInt(1)
	cnt := big.NewInt(0)

	for ; ; cnt.Add(cnt, oneIncr) {
		abort = false

		tmp = sha3.Sum256(append(cnt.Bytes(), partialHash[:]...))
		for byteNr = 0; byteNr < (uint8)(diff/8); byteNr++ {
			if tmp[byteNr] != 0 {
				abort = true
				break
			}
		}
		if abort {
			continue
		}

		if diff%8 != 0 && tmp[byteNr+1] >= 1<<(8-diff%8) {
			continue
		}
		break
	}

	return cnt
}
