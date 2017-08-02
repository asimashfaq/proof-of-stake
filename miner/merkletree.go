package miner

import (
	"golang.org/x/crypto/sha3"
)

//The merkle tree is made up of merkle nodes. It is a perfect binary tree.
type merkleNode struct {
	right, left *merkleNode
	hash        [32]byte
}

func prepareMerkleTree(txHashSlice [][32]byte) []merkleNode {

	var levelNodes []merkleNode
	var leafNode *merkleNode

	for _, txHash := range txHashSlice {
		//Construct leaf nodes
		leafNode = new(merkleNode)
		leafNode.hash = txHash
		levelNodes = append(levelNodes, *leafNode)
	}

	//We want a power of 2 for the amount of leaves for the merkle tree
	twoExp := nextTwoExponent(1, len(txHashSlice))

	//Fill up the slice for the difference of a power of 2 and the amount of hashes
	for cnt := 0; cnt < twoExp-len(txHashSlice); cnt++ {
		leafNode = new(merkleNode)
		//Make the hash of the remaining nodes the same as the last one
		leafNode.hash = levelNodes[len(levelNodes)-1].hash
		levelNodes = append(levelNodes, *leafNode)
	}

	return levelNodes
}

//Variadic functions, takes tx hashes from all tx types
func buildMerkleTree(txHashSlice ...[][32]byte) [32]byte {

	var completeSlice [][32]byte

	//Merkle root for no transactions is 0 hash
	if len(txHashSlice) == 0 {
		return [32]byte{}
	}

	//The argument is variadic, need to break down and rebuild
	for _, hashSlice := range txHashSlice {
		for _, singleHash := range hashSlice {
			completeSlice = append(completeSlice, singleHash)
		}
	}

	//If there are arguments, but all are nil, return zero hash
	if len(completeSlice) == 0 {
		return [32]byte{}
	}

	stepOver := -1
	var leftChild, rightChild, parentChild *merkleNode
	var cumulativeHash []byte

	//This function call returns the leaves of our merkle tree
	levelNodes := prepareMerkleTree(completeSlice)
	levelUpNodes := levelNodes

	//Until we just have one node which is the root node
	for len(levelUpNodes) > 1 {
		levelNodes = levelUpNodes
		levelUpNodes = []merkleNode{}

		//Loop through nodes and construct parent for left and right children
		for _, node := range levelNodes {
			stepOver++
			if stepOver%2 == 0 {
				leftChild = new(merkleNode)
				*leftChild = node
				continue
			}

			rightChild = new(merkleNode)
			*rightChild = node

			cumulativeHash = append(leftChild.hash[:], rightChild.hash[:]...)

			parentChild = new(merkleNode)
			parentChild.left = leftChild
			parentChild.right = rightChild
			parentChild.hash = sha3.Sum256(cumulativeHash)

			levelUpNodes = append(levelUpNodes, *parentChild)
		}
	}

	return levelUpNodes[0].hash
}

//We want a perfect binary tree (number of nodes exponent of two)
func nextTwoExponent(start, nrTransact int) int {
	//If there is only one tx we don't want it to be the merkle root, but being hashed with itself
	if nrTransact == 1 {
		return 2
	}
	if nrTransact == 0 {
		return 0
	}
	if start < nrTransact {
		return nextTwoExponent(start*2, nrTransact)
	}

	return start
}
