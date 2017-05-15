package bc

import (
	"golang.org/x/crypto/sha3"
)

type merkleNode struct {
	right, left *merkleNode
	hash [32]byte
}

func prepareMerkleTree(txData []fundsTx) []merkleNode {

	/*var levelNodes []merkleNode
	var parentChild *merkleNode

	for _, tx := range txData {
		//construct leaf nodes
		parentChild = new(merkleNode)
		//here we need the hash of the tx
		parentChild.hash = serializeHashContent(tx.Payload)
		levelNodes = append(levelNodes, *parentChild)
	}

	//we need power of 2 for the merkle tree
	twoExp := nextTwoExponent(1, len(txData))

	for cnt := 0; cnt < twoExp-len(txData); cnt++ {
		parentChild = new(merkleNode)
		parentChild.hash = levelNodes[len(levelNodes)-1].hash
		levelNodes = append(levelNodes, *parentChild)
	}

	return levelNodes*/
	return nil
}

func buildMerkleTree(txData []fundsTx) ([32]byte) {

	if len(txData) == 0 {
		return [32]byte{}
	}

	stepOver := -1
	var leftChild, rightChild, parentChild *merkleNode
	var cumulativeHash []byte

	levelNodes := prepareMerkleTree(txData)
	levelUpNodes := levelNodes

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

	return levelUpNodes[0].hash
}

func nextTwoExponent(start, nrTransact int) int {
	if nrTransact == 0 {
		return 0
	}
	if start < nrTransact {
		return nextTwoExponent(start*2,nrTransact)
	}
	return start
}
