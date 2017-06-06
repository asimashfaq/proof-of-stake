package bc

import (
	"fmt"
	"reflect"
)

const(
	BLOCK_REWARD = 0
	INTER_BLOCK_TIME = 60 //seconds
	BLOCKS_PER_DIFF = 5
)

var lastBlock *Block
var timestamp [BLOCKS_PER_DIFF]int64
var globalBlockCount uint64
var localBlockCount uint16
var blockDifficulty uint8

//calculation of block reward, difficulty, etc.
func getBlockReward() uint64 {

	//might get changed in the future
	return BLOCK_REWARD
}

func collectStatistics(b *Block) {
	//we need to make sure that we have the longest chain
	//long is defined as the added difficulty from the genesis block
	timestamp[localBlockCount] = b.Timestamp
	fmt.Printf("%v\n", b.Timestamp)
	globalBlockCount++
	localBlockCount++
	if localBlockCount == BLOCKS_PER_DIFF {
		calculateNewDifficulty()
		localBlockCount = 0
	}
}

func getBlockSequence(newBlock *Block) (blocksToRollback, blocksToValidate []*Block) {

	//newChainLen indicates how long the chain is to the common ancestor
	ancestor, newChainLen := findCommonAncestor(newBlock)

	if ancestor == nil {
		//ancestor not found, discard block
		return nil, nil
	}

	//we count how many blocks there are on the currently active chain
	currentChainLen := 0
	for {
		currentChainLen++
		tmpBlock := readBlock(lastBlock.PrevHash)
		if reflect.DeepEqual(tmpBlock, ancestor) {
			break
		}
	}

	//count current length and compare with new chain length
	if uint8(currentChainLen) >= newChainLen {
		//new longest chain detected, build the new chain
		blocksToValidate = append(blocksToValidate,newBlock)
		return nil,blocksToValidate
	} else {

	}

	return nil,nil
}


func findCommonAncestor(newBlock *Block) (ancestor *Block, newChainLen uint8) {

	var tmpBlock *Block
	for {
		newChainLen++
		if tmpBlock = readBlock(tmpBlock.PrevHash); tmpBlock != nil {
			//found common ancestor
			return tmpBlock, newChainLen
		}
		newBlock = tmpBlock
	}

	return nil, 0
}

func calculateNewDifficulty() {

	//BLOCKS_PER_DIFF - 1 intervals are added up
	//var intervalSum uint64

}

func getDifficulty() uint8 {
	//if chain doesn't exist yet
	if blockDifficulty == 0 {
		return 23
	}
	return blockDifficulty
}
