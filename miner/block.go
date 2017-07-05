package miner

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/lisgie/bazo_miner/protocol"
	"github.com/lisgie/bazo_miner/storage"
	"golang.org/x/crypto/sha3"
	"time"
	"github.com/lisgie/bazo_miner/p2p"
)

//acts as a temporary datastructure to fetch the payload of all transactions
type blockData struct {
	fundsTxSlice  []*protocol.FundsTx
	accTxSlice    []*protocol.AccTx
	configTxSlice []*protocol.ConfigTx
	block         *protocol.Block
}

//imitating constructor
func newBlock(prevHash [32]byte) *protocol.Block {
	b := new(protocol.Block)
	b.Header = 0x01
	b.PrevHash = prevHash
	b.StateCopy = make(map[[32]byte]*protocol.Account)
	return b
}

//this method is to validate transactions, a copy of the state
// is used for every instead of manipulating the global state
//because we the work might get interrupted by receiving a protocol.Block
func addTx(b *protocol.Block, tx protocol.Transaction) error {
	//verifies correctness for the specific transaction
	//i'd actually like to use !(&tx).verify to pass by pointer, but golang doesn't allow this
	if tx.TxFee() < FEE_MINIMUM {
		logger.Printf("Transaction fee too low: %v (minimum is: %v)\n", tx.TxFee(), FEE_MINIMUM)
		return errors.New("Transaction rejected because fee is below minimal fee threshold.")
	}

	if !verify(tx) {
		logger.Printf("Transaction could not be verified: %v\n", tx)
		return errors.New("Transaction could not be verified.")
	}

	switch tx.(type) {
	case *protocol.FundsTx:
		err := addFundsTx(b, tx.(*protocol.FundsTx))
		if err != nil {
			logger.Printf("Adding fundsTx tx failed (%v): %v\n", err, tx.(*protocol.FundsTx))
			return err
		}
	case *protocol.AccTx:
		err := addAccTx(b, tx.(*protocol.AccTx))
		if err != nil {
			logger.Printf("Adding accTx tx failed (%v): %v\n", err, tx.(*protocol.AccTx))
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

func addFundsTx(b *protocol.Block, tx *protocol.FundsTx) error {

	//I think we don't have to check for nil here as well, since this was already implicitly done with addTx(...)
	if storage.ReadClosedTx(tx.Hash()) != nil {
		return errors.New("This transaction was already included in a previous Block.")
	}

	if tx.Fee < FEE_MINIMUM {
		err := fmt.Sprintf("Fee (%v) below accepted threshold (%v)\n", tx.Fee, FEE_MINIMUM)
		return errors.New(err)
	}

	//checking if the sender account is already in the local state copy
	if _, exists := b.StateCopy[tx.From]; !exists {
		if acc := storage.State[tx.From]; acc != nil {
			hash := serializeHashContent(acc.Address)
			if hash == tx.From {
				newAcc := protocol.Account{}
				newAcc = *acc
				b.StateCopy[tx.From] = &newAcc
			}
		}
	}

	//vice versa for receiver account
	if _, exists := b.StateCopy[tx.To]; !exists {
		if acc := storage.State[tx.To]; acc != nil {
			hash := serializeHashContent(acc.Address)
			if hash == tx.To {
				newAcc := protocol.Account{}
				newAcc = *acc
				b.StateCopy[tx.To] = &newAcc
			}
		}
	}

	//rootkey doesn't need to get checked for balance
	//however, txcnt is still increased, makes things a little easiert in the state manipulation
	if !isRootKey(tx.From) {
		if (tx.Amount + tx.Fee) > b.StateCopy[tx.From].Balance {
			return errors.New("Not enough funds to complete the transaction!")
		}
	}

	//check if txcnt makes sense
	if b.StateCopy[tx.From].TxCnt != tx.TxCnt {
		err := fmt.Sprintf("Sender txCnt does not match: %v (tx.txCnt) vs. %v (state txCnt)", tx.TxCnt, b.StateCopy[tx.From].TxCnt)
		return errors.New(err)
	}

	//don't add tx if amount leads to overflow at receiver acc (amount == 0 has already been checked with verify())
	if b.StateCopy[tx.To].Balance+tx.Amount > protocol.MAX_MONEY {
		err := fmt.Sprintf("Transaction amount (%v) leads to overflow at receiver account balance (%v).\n", tx.Amount, b.StateCopy[tx.To].Balance)
		return errors.New(err)
	}

	accSender := b.StateCopy[tx.From]
	accSender.TxCnt += 1
	accSender.Balance -= tx.Amount

	accReceiver := b.StateCopy[tx.To]
	accReceiver.Balance += tx.Amount

	b.FundsTxData = append(b.FundsTxData, tx.Hash())
	storage.WriteOpenTx(tx)
	logger.Printf("Added tx to the block FundsTxData slice: %v", *tx)
	return nil
}

func addAccTx(b *protocol.Block, tx *protocol.AccTx) error {

	if storage.ReadClosedTx(tx.Hash()) != nil {
		return errors.New("This transaction was already included in a previous block.")
	}

	if tx.Fee < FEE_MINIMUM {
		err := fmt.Sprintf("Fee (%v) below accepted threshold (%v)\n", tx.Fee, FEE_MINIMUM)
		return errors.New(err)
	}

	//at this point the tx has already been verified
	accHash := sha3.Sum256(tx.PubKey[:])
	if _, exists := storage.State[accHash]; exists {
		return errors.New("Account already exists.")
	}

	b.AccTxData = append(b.AccTxData, tx.Hash())
	storage.WriteOpenTx(tx)
	logger.Printf("Added tx to the AccTxData slice: %v", *tx)
	return nil
}

func addConfigTx(b *protocol.Block, tx *protocol.ConfigTx) error {

	if storage.ReadClosedTx(tx.Hash()) != nil {
		return errors.New("This transaction was already included in a previous block.")
	}

	if tx.Fee < FEE_MINIMUM {
		err := fmt.Sprintf("Fee (%v) below accepted threshold (%v)\n", tx.Fee, FEE_MINIMUM)
		return errors.New(err)
	}

	b.ConfigTxData = append(b.ConfigTxData, tx.Hash())
	storage.WriteOpenTx(tx)
	logger.Printf("Added tx to the ConfigTxData slice: %v", *tx)
	return nil
}

func finalizeBlock(b *protocol.Block) error {
	//merkle tree only built from funds transactions
	b.MerkleRoot = buildMerkleTree(b.FundsTxData, b.AccTxData, b.ConfigTxData)
	b.Timestamp = time.Now().Unix()

	//TODO: Make this nicer, choosing by command line argument
	copy(b.Beneficiary[:], hashA[:])

	//anonymous struct
	partialHash := hashBlock(b)
	nonce, err := proofOfWork(getDifficulty(), partialHash)
	if err != nil {
		return err
	}
	b.Hash = sha3.Sum256(append(nonce.Bytes(), partialHash[:]...))

	//we need to write the proof at the end of the fixed-size byte array of length 9
	//needs to be decoded by the receiver
	for index, val := range nonce.Bytes() {
		b.Nonce[protocol.PROOF_SIZE-len(nonce.Bytes())+index] = val
	}

	//this doesn't need to be hashed, because we already have the merkle tree taking care of consistency
	b.NrFundsTx = uint16(len(b.FundsTxData))
	b.NrAccTx = uint16(len(b.AccTxData))
	b.NrConfigTx = uint8(len(b.ConfigTxData))

	return nil
}

//this function needs to be split into block syntax/PoW check and actual state change
//because there is the case that we might need to go fetch several blocks in reverse order
//and have to check the blocks first before changing the state in the correct order
func validateBlock(b *protocol.Block) error {

	blockValidation.Lock()
	defer blockValidation.Unlock()

	//TODO: Add block size check
	//this is necessary, because we need to first validate all blocks (need to fetch tx data)
	//before doing any state validation, we save all of them temporarily so we don't have to
	//refetch
	blockDataMap := make(map[[32]byte]blockData)

	blocksToRollback, blocksToValidate := getBlockSequences(b)

	if blocksToValidate == nil {
		return errors.New("Common ancestor not found or new chain shorter than current one.")
	}

	//if not the whole chain of blocks is valid, we don't consider any of them
	//this avoids the attack to create a fake long chain with only some blocks valid
	for _, block := range blocksToValidate {
		fundsTxs, accTxs, configTxs, err := preValidation(block)
		if err != nil {
			return err
		}
		blockDataMap[block.Hash] = blockData{fundsTxs, accTxs, configTxs, block}
	}

	//no rollback needed, just a new block to validate
	if len(blocksToRollback) == 0 {
		for _, block := range blocksToValidate {
			if err := stateValidation(blockDataMap[block.Hash]); err != nil {
				//if one block fails along the way, we just stop, but this is very unlikely to happen
				return err
			}
			logger.Printf("Validating block: %v\n", block)
			postValidation(blockDataMap[block.Hash])
		}
	} else {
		for _, block := range blocksToRollback {
			if err := validateBlockRollback(block); err != nil {
				return err
			}
			logger.Printf("Rolled back block: %v\n", block)
		}
		for _, block := range blocksToValidate {
			if err := stateValidation(blockDataMap[block.Hash]); err != nil {
				//if one block fails along the way, we just stop, but this is very unlikely to happen
				return err
			}
			logger.Printf("Validating block: %v\n", block)
			postValidation(blockDataMap[block.Hash])
		}
	}

	return nil
}

func preValidation(block *protocol.Block) (fundsTxSlice []*protocol.FundsTx, accTxSlice []*protocol.AccTx, configTxSlice []*protocol.ConfigTx, err error) {

	//parallel transaction data fetch
	errChan := make(chan error, 3)

	//we need to allocate slice space for the underlying array when we give pass it as reference
	fundsTxSlice = make([]*protocol.FundsTx, block.NrFundsTx)
	accTxSlice = make([]*protocol.AccTx, block.NrAccTx)
	configTxSlice = make([]*protocol.ConfigTx, block.NrConfigTx)

	go fetchFundsTxData(block, fundsTxSlice, errChan)
	go fetchAccTxData(block, accTxSlice, errChan)
	go fetchConfigTxData(block, configTxSlice, errChan)

	for cnt := 0; cnt < 3; cnt++ {
		err = <-errChan
		if err != nil {
			return nil,nil,nil,err
		}
	}

	if acc := getAccountFromHash(block.Beneficiary); acc == nil {
		return nil, nil, nil, errors.New("Beneficiary not in the State.")
	}

	startIndex := 0
	for _, singleByte := range block.Nonce {
		if singleByte != 0x00 {
			break
		}
		startIndex++
	}
	nonce := block.Nonce[startIndex:]

	partialHash := hashBlock(block)
	if block.Hash != sha3.Sum256(append(nonce, partialHash[:]...)) || !validateProofOfWork(getDifficulty(), block.Hash) {
		return nil, nil, nil, errors.New("Proof of work is incorrect.")
		logger.Println("Proof of work is incorrect.")

	}

	logger.Println("Proof of work validation passed.")

	//cmp merkle tree
	if buildMerkleTree(block.FundsTxData, block.AccTxData, block.ConfigTxData) != block.MerkleRoot {
		return nil, nil, nil, errors.New("Merkle Root incorrect.")
		logger.Println("Merkle Root incorrect.")
	}

	logger.Println("Merkle root hash passed.")
	return fundsTxSlice, accTxSlice, configTxSlice, err
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
			err := p2p.TxReq(txHash,p2p.FUNDSTX_REQ)
			if err != nil {
				errChan <- errors.New(fmt.Sprintf("FundsTx could not be read: %v", err))
				return
			}

			//blocking wait
			select {
			case fundsTx = <-p2p.FundsTxChan:
			case <-time.After(TXFETCH_TIMEOUT*time.Second):
				errChan <- errors.New("FundsTx fetch timed out.")
				return
			}

		}

		if !verifyFundsTx(fundsTx) {
			errChan <- errors.New("FundsTx could not be verified.")
			return
		}
		fundsTxSlice[cnt] = fundsTx
	}
	errChan <- nil
}

func fetchAccTxData(block *protocol.Block, accTxSlice []*protocol.AccTx, errChan chan error) {

	for cnt, txHash := range block.AccTxData {
		closedTx := storage.ReadClosedTx(txHash)
		if closedTx != nil {
			errChan <- errors.New("Block validation had accTx that was already in a previous block")
			return
		}

		var tx protocol.Transaction
		var accTx *protocol.AccTx
		tx = storage.ReadOpenTx(txHash)
		if tx != nil {
			accTx = tx.(*protocol.AccTx)
		} else {
			err := p2p.TxReq(txHash,p2p.ACCTX_REQ)
			if err != nil {
				errChan <- errors.New(fmt.Sprintf("AccTx could not be read: %v", err))
				return
			}

			//blocking wait
			select {
			case accTx = <-p2p.AccTxChan:
				//limit the waiting time to 30 seconds
			case <-time.After(TXFETCH_TIMEOUT*time.Second):
				errChan <- errors.New("AccTx fetch timed out.")
			}
		}

		if !verifyAccTx(accTx) {
			errChan <- errors.New("AccTx could not be verified.")
		}
		accTxSlice[cnt] = accTx
	}
	errChan<-nil
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
			err := p2p.TxReq(txHash,p2p.CONFIGTX_REQ)
			if err != nil {
				errChan <- errors.New(fmt.Sprintf("ConfigTx could not be read: %v", err))
				return
			}

			//blocking wait
			select {
			case configTx = <-p2p.ConfigTxChan:
				//limit the waiting time to 30 seconds
			case <-time.After(TXFETCH_TIMEOUT*time.Second):
				errChan <- errors.New("ConfigTx fetch timed out.")
				return
			}
		}

		if !verifyConfigTx(configTx) {
			errChan <- errors.New("ConfigTx could not be verified.")
			return
		}
		configTxSlice[cnt] = configTx
	}
	errChan <- nil
}

//apply to State
func stateValidation(data blockData) error {

	//we collect the fundsTx in local memory to rollback when needed
	//also, we don't want to fetch the same data several times

	//collect all fundsTx
	if err := fundsStateChange(data.fundsTxSlice); err != nil {
		return err
	}

	if err := accStateChange(data.accTxSlice); err != nil {
		//block invalid, rollback
		fundsStateChangeRollback(data.fundsTxSlice)
		return err
	}

	//can't result in an error, verify() already excluded all invalid system parameters
	//needs additionally the block hash

	//both collectTxFees as well as collectBlockReward can throw an error when the balance of the protocol overflows
	//collect fees for both transaction types
	if err := collectTxFees(data.fundsTxSlice, data.accTxSlice, data.configTxSlice, data.block.Beneficiary); err != nil {
		accStateChangeRollback(data.accTxSlice)
		fundsStateChangeRollback(data.fundsTxSlice)
		return err
	}
	//collect block reward
	if err := collectBlockReward(activeParameters.block_reward, data.block.Beneficiary); err != nil {
		collectTxFeesRollback(data.fundsTxSlice, data.accTxSlice, data.configTxSlice, data.block.Beneficiary)
		accStateChangeRollback(data.accTxSlice)
		fundsStateChangeRollback(data.fundsTxSlice)
		return err
	}

	logger.Print("Block validated and state changed accordingly: \n")
	printState()

	return nil
}

func postValidation(data blockData) {
	//put all txs from the block from open to close
	for _, tx := range data.fundsTxSlice {
		storage.WriteClosedTx(tx)
		storage.DeleteOpenTx(tx)
	}

	for _, tx := range data.accTxSlice {
		storage.WriteClosedTx(tx)
		storage.DeleteOpenTx(tx)
	}

	//block consists of system parameter changes
	for _, tx := range data.configTxSlice {
		storage.WriteClosedTx(tx)
		storage.DeleteOpenTx(tx)
	}

	//the new system parameters get active if the block was successfully validated
	configStateChange(data.configTxSlice, data.block.Hash)
	collectStatistics(data.block)

	//it might be that block is not in the openblock storage, but this doesn't matter
	storage.DeleteOpenBlock(data.block.Hash)
	storage.WriteClosedBlock(data.block)
}

func hashBlock(b *protocol.Block) (hash [32]byte) {

	var buf bytes.Buffer

	blockToHash := struct {
		prevHash    [32]byte
		header      uint8
		timestamp   int64
		merkleRoot  [32]byte
		beneficiary [32]byte
	}{
		b.PrevHash,
		b.Header,
		b.Timestamp,
		b.MerkleRoot,
		b.Beneficiary,
	}

	binary.Write(&buf, binary.BigEndian, blockToHash)
	return sha3.Sum256(buf.Bytes())
}
