package bc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"golang.org/x/crypto/sha3"
)

type merkleNode struct {
	right, left *merkleNode
	hash [32]byte
}

func buildMerkleTree(txData map[[32]byte]Transaction) (rootHash [32]byte) {

	stepOver := -1
	var leftChild, rightChild, parentChild *merkleNode
	var cumulativeHash []byte
	var levelNodes, levelUpNodes []merkleNode

	for key := range txData {
		//construct leaf nodes
		parentChild = new(merkleNode)
		parentChild.hash = txData[key]
		levelNodes = append(levelNodes, *parentChild)
	}

	levelUpNodes = levelNodes

	for len(levelUpNodes) > 1 {
		levelNodes = levelUpNodes
		levelUpNodes = []merkleNode{}
		for _, node := range levelNodes {
			stepOver++
			if stepOver%2 == 0 {
				leftChild = new(merkleNode)
				*leftChild = node
				continue
			}

			rightChild = new(merkleNode)
			*rightChild = node

			cumulativeHash = append(leftChild.hash[:],rightChild.hash[:]...)

			parentChild = new(merkleNode)
			parentChild.left = leftChild
			parentChild.right = rightChild
			parentChild.hash = sha3.Sum256(cumulativeHash)

			levelUpNodes = append(levelUpNodes, *parentChild)
		}
	}

	root := levelUpNodes[0]
	return root.hash
}

func nextTwoExponent(start, nrTransact int) int {
	if start < nrTransact {
		return nextTwoExponent(start*2,nrTransact)
	}
	return start
}

func serializeTxContent(tx TxInfo) (enc []byte) {
	// Create a struct and write it.
	var buf bytes.Buffer

	binary.Write(&buf,binary.LittleEndian, tx)

	return buf.Bytes()
}