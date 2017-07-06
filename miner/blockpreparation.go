package miner

import (
	"github.com/lisgie/bazo_miner/protocol"
	"github.com/lisgie/bazo_miner/storage"
	"sort"
)

type openTxs []protocol.Transaction

func prepareBlock(block *protocol.Block) {

	//empty mempool (opentxs)
	opentxs := storage.ReadAllOpenTxs()

	//this copy is strange, but seems to be necessary?
	//shouldn't be too bad because no deep copy
	var tmpCopy openTxs
	tmpCopy = opentxs
	sort.Sort(tmpCopy)

	for _, tx := range opentxs {
		err := addTx(block, tx)
		if err != nil {
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
		return false
	case *protocol.ConfigTx:
		return false
	}

	switch f[j].(type) {
	case *protocol.AccTx:
		return false
	case *protocol.ConfigTx:
		return false
	}
	return f[i].(*protocol.FundsTx).TxCnt < f[j].(*protocol.FundsTx).TxCnt
}
