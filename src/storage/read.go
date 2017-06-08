package storage

import (
	"fmt"
)

func ReadBlock(hash [32]byte) (encodedBlock []byte) {
	return nil
}

func ReadOpenTx(hash [32]byte) (encodedTx []byte) {

	if tx,exists := opentxs[hash]; exists {
		return tx[:]
	}
	return nil
}

func ReadClosedTx(hash [32]byte) (encodedTx []byte) {

	if tx,exists := closedtxs[hash]; exists {
		return tx[:]
	}
	return nil
}

func GetStatistics() string {
	return fmt.Sprintf("Blocks: %v\n" +
		"Unconfirmed Transactions: %v\n" +
		"Confirmed Transactions: %v\n",
		len(blocks),
		len(opentxs),
		len(closedtxs),
	)
}
