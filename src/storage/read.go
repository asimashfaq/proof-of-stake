package storage

import "fmt"

func ReadBlock(hash [32]byte) (encodedBlock []byte) {

	return nil
}

func ReadTx(hash [32]byte) (encodedTx []byte) {

	//is it a fundsTx or an accTx
	if tx,exists := txs[hash]; exists {
		return tx[:]
	}

	return nil
}

func GetStatistics() string {
	return fmt.Sprintf("State: %v\n" +
		"Root Accounts: %v\n" +
		"Blocks: %v\n" +
		"Transactions: %v\n",
		len(state),
		len(rootAccs),
		len(blocks),
		len(txs),
	)
}
