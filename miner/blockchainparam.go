package miner

import (
	"fmt"
	"github.com/lisgie/bazo_miner/protocol"
	"github.com/lisgie/bazo_miner/storage"
)

var (
	//this are "constants" that can be changed with config transactions
	//FEE_MINIMUM uint64
	//BLOCK_SIZE uint64
	//DIFF_INTERVAL uint64
	//BLOCK_INTERVAL uint64
	//BLOCK_REWARD uint64

	lastBlock        *protocol.Block
	globalBlockCount int64
	localBlockCount  int64

	target     []uint8
	targetTime *timerange
)

const (
	//in seconds
	TXFETCH_TIMEOUT    = 5
	BLOCKFETCH_TIMEOUT = 40
)

//new struct only created when at least one parameter changes in a block
type parameters struct {
	blockHash [32]byte
	//parameter
	//using int64 instead of uint64 for compatability with other programming langues
	//and also with time.Now().Unix() which is in int64
	fee_minimum    uint64
	block_size     uint64
	diff_interval  uint64
	block_interval uint64
	block_reward   uint64
	target_id uint64

	//whenever we rollback, we need to revert to that value
}

type timerange struct {
	first int64
	last  int64
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
	//Careful, this might lead to problems when run on 32-bit systems!, len(...) results an int, whose size
	// /depends on the underlying architecture

	globalBlockCount++
	localBlockCount++

	//whenever a system config change happens, we restart counting (even though it's not related to timestamping,
	//acts as "checkpointing" in some sense
	if activeParameters.blockHash == b.Hash {
		localBlockCount = 0
		targetTime = new(timerange)
		targetTime.first = b.Timestamp
	}

	if localBlockCount == int64(activeParameters.diff_interval) {
		targetTime.last = b.Timestamp
		//pre-alloation (
		target = append(target, calculateNewDifficulty(targetTime))
		logger.Printf("Target changed, new target:")
		localBlockCount = 0
		targetTime = new(timerange)
		targetTime.first = b.Timestamp
	}

	lastBlock = b
}

func collectStatisticsRollback(b *protocol.Block) {

	globalBlockCount--

	if localBlockCount == 0 {
		//localBlockCount = activeParameters.diff_interval-1

	} else {
		localBlockCount--
	}

	newLastBlock := storage.ReadClosedBlock(b.PrevHash)
	lastBlock = newLastBlock
}

func calculateNewDifficulty(t *timerange) uint8 {

	diff_now := t.last - t.first
	diff_wanted := activeParameters.block_interval * (activeParameters.diff_interval)

	diff_ratio := float32(diff_wanted) / float32(diff_now)

	//TODO: Careful, if the difference is diff_ratio is 2, we need to add (2-1) zeros, please change that as well
	//TODO: Another problem: adding 3 zeros makes it 2*2*2 = 8 times as hard, not 3 times as hard, consider the factor

	//the +-0.5 is basically the "round" function
	var target_change float32
	if diff_ratio >= 1 {
		target_change = diff_ratio
		target_change += 0.5
	} else {
		target_change = -(1/diff_ratio)
		target_change -= 0.5
	}

	//substitutes the "round" function
	fmt.Printf("diff_now = %v, diff_wanted = %v, diff_ratio = %v, target_change = %v\n", diff_now, diff_wanted, diff_ratio, target_change)

	//sanity check
	if target_change >= 4 { //make it at least 4 times as hard or easy, I think Bitcoin has something similar
		target_change = 4
	}

	return uint8(target_change + float32(target[len(target)-1]))
}

func getDifficulty() uint8 {
	return target[len(target)-1]
}
