package bc

import (
	"log"
	"errors"
)

//for blocks that already have been validated but were overwritten by a longer chain
//if this is not atomic, we're doomed
func validateBlockRollback(b *Block) error {

	fundsTxSlice, accTxSlice, err := preValidationRollback(b)
	if err != nil  {
		return err
	}

	if  err := stateValidationRollback(fundsTxSlice, accTxSlice, b.Beneficiary); err != nil {
		return err
	}

	postValidationRollback(fundsTxSlice, accTxSlice)
	deleteBlock(b.Hash)
	return nil
}

func preValidationRollback(b *Block) (fundsTxSlice []*fundsTx, accTxSlice []*accTx, err error) {

	//fetch all transactions
	for _,hash := range b.FundsTxData {
		tx := readClosedFundsTx(hash)
		if tx == nil {
			log.Printf("CRITICAL: Validated accTx was not in the confirmed tx storage: %v\n", hash)
			return nil,nil,errors.New("CRITICAL: Validated accTx was not in the confirmed tx storage")
		}
		fundsTxSlice = append(fundsTxSlice,tx)
	}

	for _,hash := range b.AccTxData {
		tx := readClosedAccTx(hash)
		if tx == nil {
			log.Printf("CRITICAL: Validated accTx was not in the confirmed tx storage: %v\n", hash)
			return nil,nil,errors.New("CRITICAL: Validated accTx was not in the confirmed tx storage")
		}
		accTxSlice = append(accTxSlice, tx)
	}

	return fundsTxSlice, accTxSlice, nil
}

func stateValidationRollback(fundsTxSlice []*fundsTx, accTxSlice []*accTx, beneficiary [32]byte) error {

	//getBlockReward needs to return a constant (same value as originally used as well)
	//the sequence is important, otherwise we subtract money from an account that does not exist anymore
	//it's exactly the opposite direction for stateValidation
	collectBlockRewardRollback(getBlockReward(),beneficiary)
	collectTxFeesRollback(fundsTxSlice, accTxSlice, beneficiary)
	accStateChangeRollback(accTxSlice)
	fundsStateChangeRollback(fundsTxSlice)
	return nil
}

func postValidationRollback(fundsTxSlice []*fundsTx, accTxSlice []*accTx) {

	//put all txs from the block from open to close
	for _,tx := range fundsTxSlice {
		hash := hashFundsTx(tx)
		writeOpenFundsTx(tx)
		deleteClosedFundsTx(hash)
	}

	for _,tx := range accTxSlice {
		hash := hashAccTx(tx)
		writeOpenAccTx(tx)
		deleteClosedAccTx(hash)
	}
}