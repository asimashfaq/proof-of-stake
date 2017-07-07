package miner

import (
	"testing"
	"fmt"
	"github.com/lisgie/bazo_miner/protocol"
)

//recognition of longer paths, common ancestor etc.
func TestTimestamps(t *testing.T) {

	cleanAndPrepare()

	//tweak parameters to test target update
	activeParameters.diff_interval = 5
	activeParameters.block_interval = 10

	prevHash := [32]byte{}
	for cnt := 0; cnt < 0; cnt++ {
		b := newBlock(prevHash)

		if cnt == 8 {
			tx,err := protocol.ConstrConfigTx(0, protocol.DIFF_INTERVAL_ID, 20,2,&RootPrivKey)
			tx2,err2 := protocol.ConstrConfigTx(0, protocol.BLOCK_INTERVAL_ID, 60, 2, &RootPrivKey)
			if err != nil || err2 != nil {
				t.Errorf("Creating config txs failed: %v, %v\n", err, err2)
			}
			err = addTx(b, tx)
			err2 = addTx(b, tx2)
			if err != nil || err2 != nil {
				t.Errorf("Adding config txs to the block failed: %v, %v\n", err, err2)
			}
		}
		finalizeBlock(b)
		validateBlock(b)
		prevHash = b.Hash

		//block is validated, check if configtx are now in the system
		if cnt == 8 {
			if activeParameters.block_interval != 60 || activeParameters.diff_interval != 20 || localBlockCount != 0 {
				t.Errorf("Block Interval: %v, Diff Interval: %v, LocalBlockCnt: %v\n",
					activeParameters.block_interval,
					activeParameters.diff_interval,
					localBlockCount,
				)
			}
		}

		fmt.Printf("Blockhash: %x, diff_interval = %v, block_interval = %v, globalCnt = %v, localCnt = %v, target: %v, targettime: %v\n",
			b.Hash[0:8],
			activeParameters.diff_interval,
			activeParameters.block_interval,
			globalBlockCount,
			localBlockCount,
			target,
			targetTime)
	}
}

func TestDifficulty(t *testing.T) {

}