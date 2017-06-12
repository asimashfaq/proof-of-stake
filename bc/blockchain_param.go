package bc

const(
	BLOCK_REWARD = 0
	INTER_BLOCK_TIME = 60 //seconds
	BLOCKS_PER_DIFF = 5 //in bitcoin, this is 2016
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
	globalBlockCount++
	localBlockCount++
	if localBlockCount == BLOCKS_PER_DIFF {
		calculateNewDifficulty()
		localBlockCount = 0
	}

	lastBlock = b
}

func collectStatisticsRollback(b *Block) {

}

func getBlockSequences(newBlock *Block) (blocksToRollback, blocksToValidate []*Block) {

	//newChainLen indicates how long the chain is to the common ancestor
	ancestor, newChain := getNewChain(newBlock)

	if ancestor == nil {
		//ancestor not found, discard block
		return nil, nil
	}

	//we count how many blocks there are on the currently active chain
	tmpBlock := lastBlock
	for {
		if tmpBlock.Hash == ancestor.Hash {
			break
		}
		blocksToRollback = append(blocksToRollback,tmpBlock)
		tmpBlock = readBlock(tmpBlock.PrevHash)
	}

	//compare current length with new chain length
	if len(blocksToRollback) >= len(newChain) {
		//current chain length is longer or equal, nothing to do
		return nil, nil
	} else {
		//new chain is longer
		return blocksToRollback, newChain
	}
}


func getNewChain(newBlock *Block) (ancestor *Block, newChain []*Block) {

	for {
		newChain = append(newChain, newBlock)

		prevBlockHash := newBlock.PrevHash
		potentialAncestor := readBlock(prevBlockHash)

		if potentialAncestor != nil {
			//found ancestor
			//we went back in time, so reverse order
			for i, j := 0, len(newChain)-1; i < j; i, j = i+1, j-1 {
				newChain[i], newChain[j] = newChain[j], newChain[i]
			}

			return potentialAncestor, newChain
		}

		//fetch the block we apparently missed
		newBlock = blockReq(prevBlockHash)
	}

	return nil, nil
}

func calculateNewDifficulty() {

	//it's smart to keep the

}

func getDifficulty() uint8 {
	//if chain doesn't exist yet
	if blockDifficulty == 0 {
		return 18
	}

	return blockDifficulty
}
