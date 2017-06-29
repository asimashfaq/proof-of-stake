package miner

import (
	"errors"
	"github.com/lisgie/bazo_miner/protocol"
	"github.com/lisgie/bazo_miner/storage"
	"log"
)

//for blocks that already have been validated but were overwritten by a longer chain
//if this is not atomic, we're doomed
func validateBlockRollback(b *protocol.Block) error {

	fundsTxSlice, accTxSlice, configTxSlice, err := preValidationRollback(b)
	if err != nil {
		return err
	}
	data := blockData{fundsTxSlice, accTxSlice, configTxSlice, b}

	//before manipulating the state, we need to go back to pre-block system parameters
	configStateChangeRollback(data.configTxSlice, b.Hash)
	if err := stateValidationRollback(data); err != nil {
		return err
	}

	postValidationRollback(data)
	return nil
}

func preValidationRollback(b *protocol.Block) (fundsTxSlice []*protocol.FundsTx, accTxSlice []*protocol.AccTx, configTxSlice []*protocol.ConfigTx, err error) {

	//fetch all transactions
	for _, hash := range b.FundsTxData {
		tx := storage.ReadClosedTx(hash).(*protocol.FundsTx)
		//this is ugly, but necessary because the encoding of the transaction throws away the full hash
		//verify acts as "enricher", e.g. writing the necessary hashes in the structure
		verify(tx)
		if tx == nil {
			log.Printf("CRITICAL: Validated fundsTx was not in the confirmed tx storage: %v\n", hash)
			return nil, nil, nil, errors.New("CRITICAL: Validated fundsTx was not in the confirmed tx storage")
		}
		fundsTxSlice = append(fundsTxSlice, tx)
	}

	for _, hash := range b.AccTxData {
		tx := storage.ReadClosedTx(hash).(*protocol.AccTx)
		if tx == nil {
			log.Printf("CRITICAL: Validated accTx was not in the confirmed tx storage: %v\n", hash)
			return nil, nil, nil, errors.New("CRITICAL: Validated accTx was not in the confirmed tx storage")
		}
		accTxSlice = append(accTxSlice, tx)
	}

	for _, hash := range b.ConfigTxData {
		tx := storage.ReadClosedTx(hash).(*protocol.ConfigTx)
		if tx == nil {
			log.Printf("CRITICAL: Validated configTx was not in the confirmed tx storage: %v\n", hash)
			return nil, nil, nil, errors.New("CRITICAL: Validated configTx was not in the confirmed tx storage")
		}
		configTxSlice = append(configTxSlice, tx)
	}

	return fundsTxSlice, accTxSlice, configTxSlice, nil
}

func stateValidationRollback(data blockData) error {

	//getBlockReward needs to return a constant (same value as originally used as well)
	//the sequence is important, otherwise we subtract money from an account that does not exist anymore
	//it's exactly the opposite direction for stateValidation

	//this has to go first, because the block that was mined, was mined with previous set system parameters
	collectBlockRewardRollback(activeParameters.block_reward, data.block.Beneficiary)
	collectTxFeesRollback(data.fundsTxSlice, data.accTxSlice, data.configTxSlice, data.block.Beneficiary)
	accStateChangeRollback(data.accTxSlice)
	fundsStateChangeRollback(data.fundsTxSlice)
	return nil
}

func postValidationRollback(data blockData) {

	//put all txs from the block from open to close
	for _, tx := range data.fundsTxSlice {
		storage.WriteOpenTx(tx)
		storage.DeleteClosedTx(tx)
	}

	for _, tx := range data.accTxSlice {
		storage.WriteOpenTx(tx)
		storage.DeleteClosedTx(tx)
	}

	for _, tx := range data.configTxSlice {
		storage.WriteOpenTx(tx)
		storage.DeleteClosedTx(tx)
	}

	collectStatisticsRollback(data.block)
	storage.DeleteBlock(data.block.Hash)
}
