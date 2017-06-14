package bc

import (
	"log"
	"errors"
)

//for blocks that already have been validated but were overwritten by a longer chain
//if this is not atomic, we're doomed
func validateBlockRollback(b *Block) error {

	fundsTxSlice, accTxSlice, configTxSlice, err := preValidationRollback(b)
	if err != nil  {
		return err
	}

	data := blockData{fundsTxSlice,accTxSlice, configTxSlice, b}

	if  err := stateValidationRollback(data); err != nil {
		return err
	}

	postValidationRollback(data)
	deleteBlock(b.Hash)
	return nil
}

func preValidationRollback(b *Block) (fundsTxSlice []*fundsTx, accTxSlice []*accTx, configTxSlice []*configTx, err error) {

	//fetch all transactions
	for _,hash := range b.FundsTxData {
		tx := readClosedFundsTx(hash)
		if tx == nil {
			log.Printf("CRITICAL: Validated fundsTx was not in the confirmed tx storage: %v\n", hash)
			return nil,nil,nil,errors.New("CRITICAL: Validated fundsTx was not in the confirmed tx storage")
		}
		fundsTxSlice = append(fundsTxSlice,tx)
	}

	for _,hash := range b.AccTxData {
		tx := readClosedAccTx(hash)
		if tx == nil {
			log.Printf("CRITICAL: Validated accTx was not in the confirmed tx storage: %v\n", hash)
			return nil,nil,nil,errors.New("CRITICAL: Validated accTx was not in the confirmed tx storage")
		}
		accTxSlice = append(accTxSlice, tx)
	}

	for _,hash := range b.ConfigTxData {
		tx := readClosedConfigTx(hash)
		if tx == nil {
			log.Printf("CRITICAL: Validated configTx was not in the confirmed tx storage: %v\n", hash)
			return nil,nil,nil,errors.New("CRITICAL: Validated configTx was not in the confirmed tx storage")
		}
		configTxSlice = append(configTxSlice,tx)
	}

	return fundsTxSlice, accTxSlice, configTxSlice, nil
}

func stateValidationRollback(data blockData) error {

	//getBlockReward needs to return a constant (same value as originally used as well)
	//the sequence is important, otherwise we subtract money from an account that does not exist anymore
	//it's exactly the opposite direction for stateValidation
	collectBlockRewardRollback(BLOCK_REWARD,data.block.Beneficiary)
	collectTxFeesRollback(data.fundsTxSlice, data.accTxSlice, data.configTxSlice, data.block.Beneficiary)
	configStateChangeRollback(data.configTxSlice)
	accStateChangeRollback(data.accTxSlice)
	fundsStateChangeRollback(data.fundsTxSlice)
	return nil
}

func postValidationRollback(data blockData) {

	//put all txs from the block from open to close
	for _,tx := range data.fundsTxSlice {
		hash := hashFundsTx(tx)
		writeOpenFundsTx(tx)
		deleteClosedFundsTx(hash)
	}

	for _,tx := range data.accTxSlice {
		hash := hashAccTx(tx)
		writeOpenAccTx(tx)
		deleteClosedAccTx(hash)
	}

	for _,tx := range data.configTxSlice {
		hash := hashConfigTx(tx)
		writeOpenConfigTx(tx)
		deleteClosedConfigTx(hash)
	}
}