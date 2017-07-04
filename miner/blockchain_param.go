package miner

import (
	"fmt"
	"github.com/lisgie/bazo_miner/protocol"
	"github.com/lisgie/bazo_miner/storage"
	"time"
	"github.com/lisgie/bazo_miner/p2p"
)

var (
	//this are "constants" that can be changed with config transactions
	FEE_MINIMUM uint64
	BLOCK_SIZE uint64
	DIFF_INTERVAL uint64
	BLOCK_INTERVAL uint64
	BLOCK_REWARD uint64

	lastBlock *protocol.Block
	globalBlockCount uint64
	localBlockCount uint64
	blockDifficulty uint8
)

const (
	//in seconds
	TXFETCH_TIMEOUT = 5
	BLOCKFETCH_TIMEOUT = 40
)

//new struct only created when at least one parameter changes in a block
type parameters struct {
	blockHash [32]byte
	//parameter
	fee_minimum    uint64
	block_size     uint64
	diff_interval  uint64
	block_interval uint64
	block_reward   uint64
}

func (param parameters) String() string {
	return fmt.Sprintf(
		"\n"+
			"Block Hash: %x\n"+
			"Block size: %v\n"+
			"Difficulty interval: %v\n"+
			"Fee minimum: %v\n"+
			"Block interval: %v\n"+
			"Block reward: %v\n",
		param.blockHash[0:8],
		param.block_size,
		param.diff_interval,
		param.fee_minimum,
		param.block_interval,
		param.block_reward,
	)
}

func collectStatistics(b *protocol.Block) {
	//we need to make sure that we have the longest chain
	//long is defined as the added difficulty from the genesis block

	//Careful, this might lead to problems when run on 32-bit systems!, len(...) results an int, whose size
	// /depends on the underlying architecture
	if uint64(len(timestamp)) <= localBlockCount {
		newTimeStamp := make([]int64, 2*(len(timestamp)+1))
		copy(newTimeStamp, timestamp)
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

func collectStatisticsRollback(b *protocol.Block) {

	globalBlockCount--
	localBlockCount--

	timestamp[int(localBlockCount)] = 0

	newLastBlock := storage.ReadClosedBlock(b.PrevHash)
	lastBlock = newLastBlock
}

func getBlockSequences(newBlock *protocol.Block) (blocksToRollback, blocksToValidate []*protocol.Block) {

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
		blocksToRollback = append(blocksToRollback, tmpBlock)
		//the block needs to be in closed storage
		tmpBlock = storage.ReadClosedBlock(tmpBlock.PrevHash)
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

func getNewChain(newBlock *protocol.Block) (ancestor *protocol.Block, newChain []*protocol.Block) {

	for {
		newChain = append(newChain, newBlock)

		prevBlockHash := newBlock.PrevHash
		potentialAncestor := storage.ReadClosedBlock(prevBlockHash)

		if potentialAncestor != nil {
			//found ancestor
			//we went back in time, so reverse order
			for i, j := 0, len(newChain)-1; i < j; i, j = i+1, j-1 {
				newChain[i], newChain[j] = newChain[j], newChain[i]
			}

			return potentialAncestor, newChain
		}

		//it might be the case that we already started a sync and the block is in the openblock storage
		newBlock = storage.ReadOpenBlock(prevBlockHash)
		if newBlock != nil {
			continue
		}
		//fetch the block we apparently missed
		p2p.BlockReq(prevBlockHash)

		//blocking wait
		select {
		case encodedBlock := <-p2p.BlockReqChan:
			newBlock = newBlock.Decode(encodedBlock)
			fmt.Printf("%v\n", newBlock)
			//limit the waiting time to 30 seconds
		case <-time.After(BLOCKFETCH_TIMEOUT*time.Second):
			return nil,nil
		}
	}

	return nil, nil
}

func calculateNewDifficulty() {

	//it's smart to keep the

}

func getDifficulty() uint8 {
	//if chain doesn't exist yet
	if blockDifficulty == 0 {
		return 10
	}

	return blockDifficulty
}
