package bc

//this are "constants" that can be changed with config transactions
var FEE_MINIMUM uint64
var BLOCK_SIZE uint64
var DIFF_INTERVAL uint64
var BLOCK_INTERVAL uint64
var BLOCK_REWARD uint64

var lastBlock *Block
var globalBlockCount uint64
var localBlockCount uint64
var blockDifficulty uint8

//new struct only created when at least one parameter changes in a block
type parameters struct {
	blockHash [32]byte
	//parameter
	fee_minimum uint64
	block_size uint64
	diff_interval uint64
	block_interval uint64
	block_reward uint64
}

func collectStatistics(b *Block) {
	//we need to make sure that we have the longest chain
	//long is defined as the added difficulty from the genesis block

	//Careful, this might lead to problems when run on 32-bit systems!, len(...) results an int, whose size
	// /depends on the underlying architecture
	if uint64(len(timestamp)) <= localBlockCount {
		newTimeStamp := make([]int64, 2*(len(timestamp)+1))
		copy(newTimeStamp,timestamp)
		timestamp = newTimeStamp
	}

	timestamp[localBlockCount] = b.Timestamp

	globalBlockCount++
	localBlockCount++


	if localBlockCount == BLOCK_INTERVAL {
		calculateNewDifficulty()
		localBlockCount = 0
	}

	lastBlock = b
}

func collectStatisticsRollback(b *Block) {

	globalBlockCount--
	localBlockCount--

	timestamp[int(localBlockCount)] = 0

	newLastBlock := readBlock(b.PrevHash)
	lastBlock = newLastBlock
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
		return 20
	}

	return blockDifficulty
}
