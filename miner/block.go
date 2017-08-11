package miner

import (
	"errors"
	"fmt"
	"github.com/lisgie/bazo_miner/p2p"
	"github.com/lisgie/bazo_miner/protocol"
	"github.com/lisgie/bazo_miner/storage"
	"golang.org/x/crypto/sha3"
	"math/big"
	"time"
)

//Datastructure to fetch the payload of all transactions, needed for state validation
type blockData struct {
	accTxSlice    []*protocol.AccTx
	fundsTxSlice  []*protocol.FundsTx
	configTxSlice []*protocol.ConfigTx
	block         *protocol.Block
}

//Block constructor, argument is the previous block in the blockchain
func newBlock(prevHash [32]byte) *protocol.Block {
	b := new(protocol.Block)
	b.PrevHash = prevHash
	b.StateCopy = make(map[[32]byte]*protocol.Account)
	return b
}

//Transaction validation operates on a copy of a tiny subset of the state (all accounts involved in transactions).
//We do not operate global state because the work might get interrupted by receiving a block that needs validation
//which is done on the global state
//b: Block to add the transaction to
//tx: A tx that implements the Transaction interface
func addTx(b *protocol.Block, tx protocol.Transaction) error {

	//activeParameters is a datastructure that stores the current system parameters, gets only changed when
	//configTxs are broadcast in the network
	if tx.TxFee() < activeParameters.fee_minimum {
		logger.Printf("Transaction fee too low: %v (minimum is: %v)\n", tx.TxFee(), activeParameters.fee_minimum)
		err := fmt.Sprintf("Transaction fee too low: %v (minimum is: %v)\n", tx.TxFee(), activeParameters.fee_minimum)
		return errors.New(err)
	}

	//There is a trade-off what tests can be made now and which have to be delayed (when dynamic state is needed
	//for inspection. The decision made is to check whether accTx and configTx have been signed with rootAcc. This
	//is a dynamic test because it needs to have access to the rootAcc state. The other option would be to include
	//the address (public key of signature) in the transaction inside the tx -> would resulted in bigger tx size.
	//So the trade-off is effectively clean abstraction vs. tx size. Everything related to fundsTx is postponed because
	//the txs are dependent on each other.
	if !verify(tx) {
		logger.Printf("Transaction could not be verified: %v\n", tx)
		return errors.New("Transaction could not be verified.")
	}

	//This check involves the state, e.g., does the account already exist, does the sender have enough balance etc.
	switch tx.(type) {
	case *protocol.AccTx:
		err := addAccTx(b, tx.(*protocol.AccTx))
		if err != nil {
			logger.Printf("Adding accTx tx failed (%v): %v\n", err, tx.(*protocol.AccTx))
			return err
		}
	case *protocol.FundsTx:
		err := addFundsTx(b, tx.(*protocol.FundsTx))
		if err != nil {
			logger.Printf("Adding fundsTx tx failed (%v): %v\n", err, tx.(*protocol.FundsTx))
			return err
		}
	case *protocol.ConfigTx:
		err := addConfigTx(b, tx.(*protocol.ConfigTx))
		if err != nil {
			logger.Printf("Adding configTx tx failed (%v): %v\n", err, tx.(*protocol.ConfigTx))
			return err
		}
	default:
		return errors.New("Transaction type not recognized.")
	}

	return nil
}

func addAccTx(b *protocol.Block, tx *protocol.AccTx) error {

	accHash := sha3.Sum256(tx.PubKey[:])
	//According to the accTx specification, we only accept new accounts _except_ if the removal bit is
	//set in the header (2nd bit)
	if tx.Header&0x02 != 0x02 {
		if _, exists := storage.State[accHash]; exists {
			return errors.New("Account already exists.")
		}
	}

	//Add the tx hash to the block header and write it to open storage (non-validated transactions)
	b.AccTxData = append(b.AccTxData, tx.Hash())
	logger.Printf("Added tx to the AccTxData slice: %v", *tx)
	return nil
}

func addFundsTx(b *protocol.Block, tx *protocol.FundsTx) error {

	//Checking if the sender account is already in the local state copy. If not and account exist, create local copy
	//If account does not exist in state, abort.
	if _, exists := b.StateCopy[tx.From]; !exists {
		if acc := storage.State[tx.From]; acc != nil {
			hash := serializeHashContent(acc.Address)
			if hash == tx.From {
				newAcc := protocol.Account{}
				newAcc = *acc
				b.StateCopy[tx.From] = &newAcc
			}
		} else {
			return errors.New(fmt.Sprintf("Sender account not present in the state: %x\n", tx.From))
		}
	}

	//Vice versa for receiver account
	if _, exists := b.StateCopy[tx.To]; !exists {
		if acc := storage.State[tx.To]; acc != nil {
			hash := serializeHashContent(acc.Address)
			if hash == tx.To {
				newAcc := protocol.Account{}
				newAcc = *acc
				b.StateCopy[tx.To] = &newAcc
			}
		} else {
			return errors.New(fmt.Sprintf("Receiver account not present in the state: %x\n", tx.From))
		}
	}

	//Root accounts are exempt from balance requirements. All other accounts need to have (at least)
	//fee + amount to spend as balance available
	if !isRootKey(tx.From) {
		if (tx.Amount + tx.Fee) >= b.StateCopy[tx.From].Balance {
			return errors.New("Not enough funds to complete the transaction!")
		}
	}

	//Transaction count need to match the state, preventing replay attacks
	if b.StateCopy[tx.From].TxCnt != tx.TxCnt {
		err := fmt.Sprintf("Sender txCnt does not match: %v (tx.txCnt) vs. %v (state txCnt)", tx.TxCnt, b.StateCopy[tx.From].TxCnt)
		return errors.New(err)
	}

	//Prevent balance overflow in receiver account
	if b.StateCopy[tx.To].Balance+tx.Amount > MAX_MONEY {
		err := fmt.Sprintf("Transaction amount (%v) leads to overflow at receiver account balance (%v).\n", tx.Amount, b.StateCopy[tx.To].Balance)
		return errors.New(err)
	}

	//Update state copy
	accSender := b.StateCopy[tx.From]
	accSender.TxCnt += 1
	accSender.Balance -= tx.Amount

	accReceiver := b.StateCopy[tx.To]
	accReceiver.Balance += tx.Amount

	//Add the tx hash to the block header and write it to open storage (non-validated transactions)
	b.FundsTxData = append(b.FundsTxData, tx.Hash())
	logger.Printf("Added tx to the block FundsTxData slice: %v", *tx)
	return nil
}

func addConfigTx(b *protocol.Block, tx *protocol.ConfigTx) error {

	//No further checks needed, static checks were already done with verify()
	b.ConfigTxData = append(b.ConfigTxData, tx.Hash())
	logger.Printf("Added tx to the ConfigTxData slice: %v", *tx)
	return nil
}

//This function prepares the block to broadcast into the network. No new txs are added at this point.
func finalizeBlock(b *protocol.Block) error {

	//Merkle tree includes the hashes of all txs
	b.MerkleRoot = buildMerkleTree(b.AccTxData, b.FundsTxData, b.ConfigTxData)

	b.Timestamp = time.Now().Unix()

	//BENEFICIARY is a config parameter set in config.go
	beneficiary, _ := new(big.Int).SetString(BENEFICIARY, 16)
	copy(b.Beneficiary[:], beneficiary.Bytes())

	partialHash := b.HashBlock()
	nonce, err := proofOfWork(getDifficulty(), partialHash, b.PrevHash)
	if err != nil {
		return err
	}
	b.Nonce = nonce
	//Put pieces to gether to get the final hash
	b.Hash = sha3.Sum256(append(nonce[:], partialHash[:]...))

	//This doesn't need to be hashed, because we already have the merkle tree taking care of consistency

	b.NrAccTx = uint16(len(b.AccTxData))
	b.NrFundsTx = uint16(len(b.FundsTxData))
	b.NrConfigTx = uint8(len(b.ConfigTxData))

	return nil
}

//This function is split into block syntax/PoW check and actual state change
//because there is the case that we might need to go fetch several blocks
// and have to check the blocks first before changing the state in the correct order
func validateBlock(b *protocol.Block) error {

	//This mutex is necessary that own-mined blocks and received blocks from the network are not
	//validated concurrently
	blockValidation.Lock()
	defer blockValidation.Unlock()

	//Prepare datastructure to fill tx payloads
	blockDataMap := make(map[[32]byte]blockData)

	//Get the right branch, and a list of blocks to rollback (if necessary)
	blocksToRollback, blocksToValidate := getBlockSequences(b)

	//Verify block time is dynamic and corresponds to system time at the time of retrieval.
	//If we're syncing or far behind, we cannot do this dynamic check
	//We therefore include a boolean uptodate. If it's true we consider ourselves uptodate and
	//do dynamic time checking
	if len(blocksToValidate) > DELAYED_BLOCKS {
		uptodate = false
	} else {
		uptodate = true
	}

	if blocksToValidate == nil {
		return errors.New("Common ancestor not found or new chain shorter than current one.")
	}

	//If not the whole chain of blocks is valid, we don't do state changes on any of them before
	//making sure they're properly formed. This avoids the attack to create a fake long chain with
	//only some blocks valid
	for _, block := range blocksToValidate {
		//Fetching payload data from the txs (if necessary, ask other miners)
		accTxs, fundsTxs, configTxs, err := preValidation(block)
		if err != nil {
			return err
		}
		blockDataMap[block.Hash] = blockData{accTxs, fundsTxs, configTxs, block}
	}

	//No rollback needed, just a new block to validate
	if len(blocksToRollback) == 0 {
		for _, block := range blocksToValidate {
			if err := stateValidation(blockDataMap[block.Hash]); err != nil {
				return err
			}
			logger.Printf("Validating block: %vState:\n%v", block, getState())
			postValidation(blockDataMap[block.Hash])
		}
	} else {
		for _, block := range blocksToRollback {
			if err := validateBlockRollback(block); err != nil {
				return err
			}
			logger.Printf("Rolled back block: %vState:\n%v", block, getState())
		}
		for _, block := range blocksToValidate {
			if err := stateValidation(blockDataMap[block.Hash]); err != nil {
				return err
			}
			logger.Printf("Validating block: %vState:\n%v",block, getState())
			postValidation(blockDataMap[block.Hash])
		}
	}

	return nil
}

//Doesn't involve any state changes
func preValidation(block *protocol.Block) (accTxSlice []*protocol.AccTx, fundsTxSlice []*protocol.FundsTx, configTxSlice []*protocol.ConfigTx, err error) {

	//This dynamic check is only done if we're up-to-date with syncing. Otherwise, timestamp is not checked
	//Other miners (which are up-to-date) made sure that this is correct
	if uptodate {
		if err := timestampCheck(block.Timestamp); err != nil {
			return nil, nil, nil, err
		}
	}

	if block.GetSize() > activeParameters.block_size {
		return nil, nil, nil, errors.New("Block size too large.")
	}

	//Duplicates are not allowed, use tx hash hasmap to easily check for duplicates
	duplicates := make(map[[32]byte]bool)
	for _, txHash := range block.AccTxData {
		if _, exists := duplicates[txHash]; exists {
			return nil, nil, nil, errors.New("Duplicate Transaction Hash detected.")
		}
		duplicates[txHash] = true
	}
	for _, txHash := range block.FundsTxData {
		if _, exists := duplicates[txHash]; exists {
			return nil, nil, nil, errors.New("Duplicate Transaction Hash detected.")
		}
		duplicates[txHash] = true
	}
	for _, txHash := range block.ConfigTxData {
		if _, exists := duplicates[txHash]; exists {
			return nil, nil, nil, errors.New("Duplicate Transaction Hash detected.")
		}
		duplicates[txHash] = true
	}

	//We fetch tx data for each type in parallel -> performance boost
	errChan := make(chan error, 3)

	//we need to allocate slice space for the underlying array when we pass them as reference
	accTxSlice = make([]*protocol.AccTx, block.NrAccTx)
	fundsTxSlice = make([]*protocol.FundsTx, block.NrFundsTx)
	configTxSlice = make([]*protocol.ConfigTx, block.NrConfigTx)

	go fetchAccTxData(block, accTxSlice, errChan)
	go fetchFundsTxData(block, fundsTxSlice, errChan)
	go fetchConfigTxData(block, configTxSlice, errChan)

	//Wait for all goroutines to finish
	for cnt := 0; cnt < 3; cnt++ {
		err = <-errChan
		if err != nil {
			return nil, nil, nil, err
		}
	}

	//Does the beneficiary exist in the state
	if acc := storage.GetAccountFromHash(block.Beneficiary); acc == nil {
		return nil, nil, nil, errors.New("Beneficiary not in the State.")
	}

	//PoW validation
	partialHash := block.HashBlock()
	if block.Hash != sha3.Sum256(append(block.Nonce[:], partialHash[:]...)) || !validateProofOfWork(getDifficulty(), block.Hash) {
		return nil, nil, nil, errors.New("Proof of work is incorrect.")
		logger.Println("Proof of work is incorrect.")

	}

	//Merkle Tree validation
	if buildMerkleTree(block.AccTxData, block.FundsTxData, block.ConfigTxData) != block.MerkleRoot {
		return nil, nil, nil, errors.New("Merkle Root incorrect.")
		logger.Println("Merkle Root incorrect.")
	}

	return accTxSlice, fundsTxSlice, configTxSlice, err
}


//Only blocks with timestamp not diverging from system time (past or future) more than one hour are accepted
func timestampCheck(timestamp int64) error {
	systemTime := p2p.ReadSystemTime()
	if timestamp > systemTime {
		if timestamp-systemTime > int64(time.Hour.Seconds()) {
			return errors.New("Timestamp was too far in the future.\n")
		}
	} else {
		if systemTime-timestamp > int64(time.Hour.Seconds()) {
			return errors.New("Timestamp was too far in the past.\n")
		}
	}
	return nil
}

//We use slices (not maps) because order is now important
func fetchAccTxData(block *protocol.Block, accTxSlice []*protocol.AccTx, errChan chan error) {

	for cnt, txHash := range block.AccTxData {
		//Reject blocks that have txs which have already been validated
		closedTx := storage.ReadClosedTx(txHash)
		if closedTx != nil {
			errChan <- errors.New("Block validation had accTx that was already in a previous block")
			return
		}

		var tx protocol.Transaction
		var accTx *protocol.AccTx
		//Tx is either in open storage or needs to be fetched from the network
		tx = storage.ReadOpenTx(txHash)
		if tx != nil {
			accTx = tx.(*protocol.AccTx)
		} else {
			err := p2p.TxReq(txHash, p2p.ACCTX_REQ)
			if err != nil {
				errChan <- errors.New(fmt.Sprintf("AccTx could not be read: %v", err))
				return
			}

			//Blocking Wait
			select {
			case accTx = <-p2p.AccTxChan:
				//Limit the waiting time for TXFETCH_TIMEOUT seconds
			case <-time.After(TXFETCH_TIMEOUT * time.Second):
				errChan <- errors.New("AccTx fetch timed out.")
			}
		}

		accTxSlice[cnt] = accTx
	}
	errChan <- nil
}

func fetchFundsTxData(block *protocol.Block, fundsTxSlice []*protocol.FundsTx, errChan chan error) {

	for cnt, txHash := range block.FundsTxData {
		closedTx := storage.ReadClosedTx(txHash)
		if closedTx != nil {
			errChan <- errors.New("Block validation had fundsTx that was already in a previous block")
			return
		}

		var tx protocol.Transaction
		var fundsTx *protocol.FundsTx
		tx = storage.ReadOpenTx(txHash)
		if tx != nil {
			fundsTx = tx.(*protocol.FundsTx)
		} else {
			err := p2p.TxReq(txHash, p2p.FUNDSTX_REQ)
			if err != nil {
				errChan <- errors.New(fmt.Sprintf("FundsTx could not be read: %v", err))
				return
			}

			select {
			case fundsTx = <-p2p.FundsTxChan:
			case <-time.After(TXFETCH_TIMEOUT * time.Second):
				errChan <- errors.New("FundsTx fetch timed out.")
				return
			}
		}

		fundsTxSlice[cnt] = fundsTx
	}
	errChan <- nil
}

func fetchConfigTxData(block *protocol.Block, configTxSlice []*protocol.ConfigTx, errChan chan error) {

	for cnt, txHash := range block.ConfigTxData {
		closedTx := storage.ReadClosedTx(txHash)
		if closedTx != nil {
			errChan <- errors.New("Block validation had configTx that was already in a previous block")
			return
		}

		var tx protocol.Transaction
		var configTx *protocol.ConfigTx
		tx = storage.ReadOpenTx(txHash)
		if tx != nil {
			configTx = tx.(*protocol.ConfigTx)
		} else {
			err := p2p.TxReq(txHash, p2p.CONFIGTX_REQ)
			if err != nil {
				errChan <- errors.New(fmt.Sprintf("ConfigTx could not be read: %v", err))
				return
			}

			select {
			case configTx = <-p2p.ConfigTxChan:
			case <-time.After(TXFETCH_TIMEOUT * time.Second):
				errChan <- errors.New("ConfigTx fetch timed out.")
				return
			}
		}

		configTxSlice[cnt] = configTx
	}
	errChan <- nil
}

//Dynamic state check
func stateValidation(data blockData) error {

	//The sequence of validation matters. If we start with accs, then fund transfers can be done in the same block
	//even though the accounts did not exist before the block validation
	if err := accStateChange(data.accTxSlice); err != nil {
		return err
	}

	if err := fundsStateChange(data.fundsTxSlice); err != nil {
		accStateChangeRollback(data.accTxSlice)
		return err
	}

	if err := collectTxFees(data.accTxSlice, data.fundsTxSlice, data.configTxSlice, data.block.Beneficiary); err != nil {
		fundsStateChangeRollback(data.fundsTxSlice)
		accStateChangeRollback(data.accTxSlice)
		return err
	}

	if err := collectBlockReward(activeParameters.block_reward, data.block.Beneficiary); err != nil {
		collectTxFeesRollback(data.accTxSlice, data.fundsTxSlice, data.configTxSlice, data.block.Beneficiary)
		fundsStateChangeRollback(data.fundsTxSlice)
		accStateChangeRollback(data.accTxSlice)
		return err
	}

	return nil
}

func postValidation(data blockData) {
	//Write all open transactions to closed/validated storage
	for _, tx := range data.accTxSlice {
		storage.WriteClosedTx(tx)
		storage.DeleteOpenTx(tx)
	}

	for _, tx := range data.fundsTxSlice {
		storage.WriteClosedTx(tx)
		storage.DeleteOpenTx(tx)
	}

	for _, tx := range data.configTxSlice {
		storage.WriteClosedTx(tx)
		storage.DeleteOpenTx(tx)
	}

	//The new system parameters get active if the block was successfully validated
	//This is done after state validation (in contrast to accTx/fundsTx).
	//Conversely, if blocks are rolled back, the system parameters are changed first
	configStateChange(data.configTxSlice, data.block.Hash)
	//Collects meta information about the block (and handled difficulty adaption)
	collectStatistics(data.block)

	//It might be that block is not in the openblock storage, but this doesn't matter
	storage.DeleteOpenBlock(data.block.Hash)
	storage.WriteClosedBlock(data.block)
}
