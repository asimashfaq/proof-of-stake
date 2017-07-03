package miner

import (
	"errors"
	"golang.org/x/crypto/sha3"
	"math/big"
	"time"
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

func proofOfWork(diff uint8, partialHash [32]byte) (*big.Int, error) {

	logger.Printf("Start mining a new block with difficulty: %v\n", diff)

	var tmp [32]byte
	var byteNr uint8
	var abort bool
	//big int needed because int64 overflows if nonce too large
	oneIncr := big.NewInt(1)
	cnt := big.NewInt(0)

	startedWith := lastBlock.Hash

	for ; ; cnt.Add(cnt, oneIncr) {

		//CPU IS BUUUUUUUUUUUUUUURNING otherwise
		time.Sleep(3 * time.Millisecond)

		if startedWith != lastBlock.Hash {
			return nil, errors.New("Abort mining, another block has been successfully validated in the meantime")
		}
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

	return cnt, nil
}
