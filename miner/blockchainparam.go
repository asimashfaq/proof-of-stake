package miner

import (
	"fmt"
	"github.com/lisgie/bazo_miner/protocol"
	"github.com/lisgie/bazo_miner/storage"
	"math"
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
	currentTargetTime *timerange
)

const (
	//in seconds
	TXFETCH_TIMEOUT    = 5
	BLOCKFETCH_TIMEOUT = 40
)

//new struct only created when at least one parameter changes in a block
type parameters struct {
	blockHash [32]byte

	fee_minimum    uint64
	block_size     uint64
	diff_interval  uint64
	block_interval uint64
	block_reward   uint64
	target_id uint64
}

type timerange struct {
	first int64
	last  int64
}

//we need to store the history or timeranges to revert in case of rollbacks
var targetTimes []timerange

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

	globalBlockCount++
	localBlockCount++

	//TODO:
	//there are three system parameters that influence target caclculation, namely diffinterval and blockinterval
	//and target it self. If one of these params have been changed, the calculation starts anew.

	//For now, just change everything in case of a parameter change
	if activeParameters.blockHash == b.Hash {
		//localBlockCount = 0
		//currentTargetTime = new(timerange)
		//currentTargetTime.first = b.Timestamp
	}

	if localBlockCount == int64(activeParameters.diff_interval) {

		currentTargetTime.last = b.Timestamp
		//The genesis block has timestamp = 0. This simplifies certain things: Every miner can start with an already
		//existing genesis block (because all fields are set to 0). The "find common ancestor" algorithm can then
		//use the genesis block as a common ancestor for new miners who have not synchronized with the chain yet.
		if currentTargetTime.first == 0 {
			target = append(target, target[len(target)-1])
			fmt.Printf("Genesis: %v\n", target)
		} else {
			target = append(target, calculateNewDifficulty(currentTargetTime))
			fmt.Printf("Target update: %v\n", target)

		}

		targetTimes = append(targetTimes, *currentTargetTime)
		fmt.Printf("Target times: %v\n", targetTimes)

		logger.Printf("Target changed, new target: %v", target[len(target)-1])
		localBlockCount = 0
		currentTargetTime = new(timerange)
		currentTargetTime.first = b.Timestamp
	}

	lastBlock = b
}

func collectStatisticsRollback(b *protocol.Block) {

	globalBlockCount--

	//never rollback the genesis blocks
	if localBlockCount == 0 && globalBlockCount != 0 {
		localBlockCount = int64(activeParameters.diff_interval)-1
		//target rollback
		fmt.Printf("Before: %v\n", target)
		target = target[:len(target)-1]
		fmt.Printf("After: %v\n", target)
		currentTargetTime.first = targetTimes[len(targetTimes)-1].first
		fmt.Printf("Before: %v\n", targetTimes)
		targetTimes = targetTimes[:len(targetTimes)-1]
		fmt.Printf("After: %v\n", targetTimes)
	} else {
		localBlockCount--
	}

	lastBlock = storage.ReadClosedBlock(b.PrevHash)
}

func calculateNewDifficulty(t *timerange) uint8 {

	diff_now := t.last - t.first
	diff_wanted := activeParameters.block_interval * (activeParameters.diff_interval)

	diff_ratio := float64(diff_wanted) / float64(diff_now)

	//if the last is earlier time than first, we get a negative number, can't take the log from that
	//this precipitates that reasonable parameter should be chosen for block interval/diff interval
	//such that this case does not happen. In case it still does, we give the current difficulty back
	if diff_ratio < 0 {
		return getDifficulty()
	}
	target_change := math.Log2(diff_ratio)

	//the +-0.5 is basically the "round" function
	if target_change > 0 {
		target_change += 0.5
	} else if target_change < 0 {
		target_change -= 0.5
	}

	//Sanity check! Make it at most 3 times as hard or easy, Bitcoin has a similar sanity
	if target_change > 3 {
		target_change = 3
	} else if target_change < -3 {
		target_change = -3
	}

	//Rounding down (for positive values) and runding up (for negative values)
	target_change_rounded := uint8(target_change)

	//substitutes the "round" function
	return target_change_rounded + target[len(target)-1]
}

func getDifficulty() uint8 {
	return target[len(target)-1]
}
