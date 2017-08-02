package miner

import (
	"github.com/lisgie/bazo_miner/protocol"
	"github.com/lisgie/bazo_miner/storage"
	"sort"
)

//The code here is needed if a new block is being built. All open (not yet validated) transactions are first fetched
//from the mempool and then sorted. The sorting is important because if transactions are fetched from the mempool
//they're received in random order (because it's implemented as map). However, if a user wants to issue more fundsTxs
//they need to be sorted according to increasing txCnt, this greatly increases throughput.

type openTxs []protocol.Transaction

func prepareBlock(block *protocol.Block) {

	//Fetch all txs from mempool (opentxs)
	opentxs := storage.ReadAllOpenTxs()

	//This copy is strange, but seems to be necessary to leverage the sort interface.
	//Shouldn't be too bad because no deep copy.
	var tmpCopy openTxs
	tmpCopy = opentxs

	sort.Sort(tmpCopy)
	for _, tx := range opentxs {
		//Prevent block size overflow
		if block.GetSize() + tx.Size() > activeParameters.block_size {
			break
		}
		err := addTx(block, tx)
		if err != nil {
			//If the tx is invalid, we remove it completely, prevents starvation in the mempool
			storage.DeleteOpenTx(tx)
		}
	}
}

//Implement the sort interface
func (f openTxs) Len() int {
	return len(f)
}

func (f openTxs) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func (f openTxs) Less(i, j int) bool {
	//comparison only makes sense if both tx are fundstxs
	//why can we only do that with switch, and not e.g., if tx.(type) == ..?
	switch f[i].(type) {
	case *protocol.AccTx:
		//We only want to sort a subset of all transactions, namely all fundstxs.
		//However, to successfully do that we have to place all other txs at the beginning
		//The order between accTxs and configTxs doesn't matter
		return true
	case *protocol.ConfigTx:
		return true
	}

	switch f[j].(type) {
	case *protocol.AccTx:
		return false
	case *protocol.ConfigTx:
		return false
	}

	return f[i].(*protocol.FundsTx).TxCnt < f[j].(*protocol.FundsTx).TxCnt
}
