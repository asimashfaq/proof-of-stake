package miner

import (
	"fmt"
	"github.com/lisgie/bazo_miner/protocol"
	"github.com/lisgie/bazo_miner/storage"
	"math"
)

var (
	lastBlock        *protocol.Block
	globalBlockCount = int64(-1)
	localBlockCount  = int64(-1)
	target            []uint8 //Stores the history of target values
	currentTargetTime *timerange //Corresponds to the active timerange
)

//An instance of this datastructure is created whenever there are system parameter changes
//The blockhash is additionally recorded to know which blocks the parameter change belongs to.
//This is necessary, because the system records ALL config txs (even those who have no corresponding
//code to execute [e.g., when they're running an older version of the code]).
type parameters struct {
	blockHash [32]byte

	fee_minimum    uint64
	block_size     uint64
	diff_interval  uint64
	block_interval uint64
	block_reward   uint64
}

//Captures first and last timestamp of the intended blocks of the range
type timerange struct {
	first int64
	last  int64
}

//We need to store the history or timeranges to revert in case of rollbacks
var targetTimes []timerange

func collectStatistics(b *protocol.Block) {

	globalBlockCount++
	localBlockCount++

	if localBlockCount == int64(activeParameters.diff_interval) {

		currentTargetTime.last = b.Timestamp
		//The genesis block has timestamp = 0. This simplifies certain things: Every miner can start with an already
		//existing genesis block (because all fields are set to 0). The "find common ancestor" algorithm can then
		//use the genesis block as a common ancestor for new miners who have not synchronized with the chain yet.
		if currentTargetTime.first == 0 {
			target = append(target, target[len(target)-1])
		} else {
			target = append(target, calculateNewDifficulty(currentTargetTime))
		}

		targetTimes = append(targetTimes, *currentTargetTime)

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
		localBlockCount = int64(activeParameters.diff_interval) - 1
		//Target rollback
		target = target[:len(target)-1]
		currentTargetTime.first = targetTimes[len(targetTimes)-1].first
		targetTimes = targetTimes[:len(targetTimes)-1]
	} else {
		localBlockCount--
	}

	lastBlock = storage.ReadClosedBlock(b.PrevHash)
}

func calculateNewDifficulty(t *timerange) uint8 {

	//Time difference between the first and last block in the measured range
	diff_now := t.last - t.first
	//This is how long it should have taken
	diff_wanted := activeParameters.block_interval * (activeParameters.diff_interval)

	diff_ratio := float64(diff_wanted) / float64(diff_now)

	//If the last is earlier time than first, we get a negative number, can't take the log from that
	//this precipitates that reasonable parameter should be chosen for block interval/diff interval
	//such that this case does not happen. In case it still does, we give the current difficulty back
	if diff_ratio < 0 {
		return getDifficulty()
	}

	//Take the log2 from the diff_ratio, because adding a zero makes it twice as hard, adding two zeros four times as
	//hard etc.
	target_change := math.Log2(diff_ratio)

	//the +-0.5 is basically the "round" function
	if target_change > 0 {
		target_change += 0.5
	} else if target_change < 0 {
		target_change -= 0.5
	}

	//Sanity check! Make it at most 3 times as hard or easy, Bitcoin has a similar check
	if target_change > 3 {
		target_change = 3
	} else if target_change < -3 {
		target_change = -3
	}

	//Rounding down (for positive values) and runding up (for negative values)
	target_change_rounded := uint8(target_change)

	//Return the new target based on the calculation and the current target
	return target_change_rounded + target[len(target)-1]
}

func getDifficulty() uint8 {
	return target[len(target)-1]
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