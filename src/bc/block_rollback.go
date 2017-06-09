package bc

import (
	"log"
	"errors"
)

//for blocks that already have been validated but were overwritten by a longer chain
//if this is not atomic, we're doomed
func blockRollback(b *Block) error {

	var fundsTxSlice []*fundsTx
	var accTxSlice []*accTx
	//fetch all transactions
	for _,hash := range b.FundsTxData {
		tx := readClosedFundsTx(hash)
		if tx == nil {
			log.Printf("CRITICAL: Validated accTx was not in the confirmed tx storage: %v\n", hash)
			return errors.New("CRITICAL: Validated accTx was not in the confirmed tx storage")
		}
		fundsTxSlice = append(fundsTxSlice,tx)

		//switch from confirmed to unconfirmed
		deleteClosedFundsTx(hash)
		writeOpenFundsTx(tx)
	}

	for _,hash := range b.AccTxData {
		tx := readClosedAccTx(hash)
		if tx == nil {
			log.Printf("CRITICAL: Validated accTx was not in the confirmed tx storage: %v\n", hash)
			return errors.New("CRITICAL: Validated accTx was not in the confirmed tx storage")
		}
		accTxSlice = append(accTxSlice, tx)

		deleteClosedAccTx(hash)
		writeOpenAccTx(tx)
	}
	return nil
}

func stateValidationRollback(b *Block) {

}

func postValidationRollback(b *Block) {

}
