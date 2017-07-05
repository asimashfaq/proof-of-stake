package miner

import (
	"errors"
	"github.com/lisgie/bazo_miner/protocol"
	"github.com/lisgie/bazo_miner/storage"
)

//for blocks that already have been validated but were overwritten by a longer chain
//if this is not atomic, we're doomed
func validateBlockRollback(b *protocol.Block) error {

	accTxSlice, fundsTxSlice, configTxSlice, err := preValidationRollback(b)
	if err != nil {
		return err
	}
	data := blockData{accTxSlice, fundsTxSlice, configTxSlice, b}

	//before manipulating the state, we need to go back to pre-block system parameters
	configStateChangeRollback(data.configTxSlice, b.Hash)
	if err := stateValidationRollback(data); err != nil {
		return err
	}

	postValidationRollback(data)
	return nil
}

func preValidationRollback(b *protocol.Block) (accTxSlice []*protocol.AccTx, fundsTxSlice []*protocol.FundsTx, configTxSlice []*protocol.ConfigTx, err error) {

	//fetch all transactions from closed storage
	for _, hash := range b.AccTxData {
		var accTx *protocol.AccTx
		tx := storage.ReadClosedTx(hash)
		if tx == nil {
			logger.Printf("CRITICAL: Validated accTx was not in the confirmed tx storage: %v\n", hash)
			return nil, nil, nil, errors.New("CRITICAL: Validated accTx was not in the confirmed tx storage")
		} else {
			accTx = tx.(*protocol.AccTx)
		}
		accTxSlice = append(accTxSlice, accTx)
	}

	for _, hash := range b.FundsTxData {
		var fundsTx *protocol.FundsTx
		tx := storage.ReadClosedTx(hash)
		if tx == nil {
			logger.Printf("CRITICAL: Validated fundsTx was not in the confirmed tx storage: %v\n", hash)
			return nil, nil, nil, errors.New("CRITICAL: Validated fundsTx was not in the confirmed tx storage")
		} else {
			fundsTx = tx.(*protocol.FundsTx)
		}
		fundsTxSlice = append(fundsTxSlice, fundsTx)
	}

	for _, hash := range b.ConfigTxData {
		var configTx *protocol.ConfigTx
		tx := storage.ReadClosedTx(hash)
		if tx == nil {
			logger.Printf("CRITICAL: Validated configTx was not in the confirmed tx storage: %v\n", hash)
			return nil, nil, nil, errors.New("CRITICAL: Validated configTx was not in the confirmed tx storage")
		} else {
			configTx = tx.(*protocol.ConfigTx)
		}
		configTxSlice = append(configTxSlice, configTx)
	}

	return accTxSlice, fundsTxSlice, configTxSlice, nil
}

func stateValidationRollback(data blockData) error {

	//getBlockReward needs to return a constant (same value as originally used as well)
	//the sequence is important, otherwise we subtract money from an account that does not exist anymore
	//it's exactly the opposite direction for stateValidation

	//this has to go first, because the block that was mined, was mined with previous set system parameters
	collectBlockRewardRollback(activeParameters.block_reward, data.block.Beneficiary)
	collectTxFeesRollback(data.accTxSlice, data.fundsTxSlice, data.configTxSlice, data.block.Beneficiary)
	fundsStateChangeRollback(data.fundsTxSlice)
	accStateChangeRollback(data.accTxSlice)
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

	//for transactions we switch from closed to open. However, we do not write back blocks
	//to open storage, because in case of rollback the chain they belonged to is likely to starve
	storage.DeleteClosedBlock(data.block.Hash)
}
