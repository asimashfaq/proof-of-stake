package miner

import (
	"github.com/lisgie/bazo_miner/p2p"
	"github.com/lisgie/bazo_miner/protocol"
	"github.com/lisgie/bazo_miner/storage"
	"time"
)

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
			//limit the waiting time to 30 seconds
		case <-time.After(BLOCKFETCH_TIMEOUT * time.Second):
			return nil, nil
		}
	}

	return nil, nil
}
