package bc

import (
	"golang.org/x/crypto/sha3"
	"time"
	"errors"
	"log"
	"fmt"
	"bytes"
	"encoding/binary"
)

const (
	HASH_LEN = 32
	PROOF_SIZE = 9
	BLOCKHEADER_SIZE = 151
)

type transaction interface {
	verify() bool
}

//acts as a temporary datastructure to fetch the payload of all transactions
type blockData struct{
	fundsTxSlice []*fundsTx
	accTxSlice []*accTx
	configTxSlice []*configTx
	block *Block
}

type Block struct {
	Header byte
	Hash [32]byte
	PrevHash [32]byte
	Nonce [PROOF_SIZE]byte //72-bit, enough even if the network gets really large
	Timestamp int64
	MerkleRoot [32]byte
	Beneficiary [32]byte
	NrFundsTx uint16
	NrAccTx uint16
	NrConfigTx uint8
	//this field will not be exported, this is just to avoid race conditions for the global state
	stateCopy map[[32]byte]*Account
	FundsTxData [][32]byte
	AccTxData [][32]byte
	ConfigTxData [][32]byte
}

//imitating constructor
func newBlock() *Block {
	b := Block{}
	b.Header = 0x01
	b.stateCopy = make(map[[32]byte]*Account)
	return &b
}

//this method is to validate transactions, a copy of the state
// is used for every instead of manipulating the global state
//because we the work might get interrupted by receiving a block
func (b *Block) addTx(tx transaction) error {
	//verifies correctness for the specific transaction
	//i'd actually like to use !(&tx).verify to pass by pointer, but golang doesn't allow this
	if !(tx).verify() {
		log.Printf("Transaction could not be verified: %v\n", tx)
		return errors.New("Transaction could not be verified.")
	}

	//it would be much nicer if we could do a minimal fee check here, but this isn't so easy
	//with the lack of OOP support from golang. It would require to access "parent data" (tx.Fee),
	//so I'll just check for both FundsTx and AccTx
	switch tx.(type) {
	case *fundsTx:
		err := b.addFundsTx(tx.(*fundsTx))
		if err != nil {
			log.Printf("Adding fundsTx tx failed (%v): %v\n",err, tx.(*fundsTx))
			return err
		}
	case *accTx:
		err := b.addAccTx(tx.(*accTx))
		if err != nil {
			log.Printf("Adding accTx tx failed (%v): %v\n", err,tx.(*accTx))
			return err
		}
	case *configTx:
		err := b.addConfigTx(tx.(*configTx))
		if err != nil {
			log.Printf("Adding configTx tx failed (%v): %v\n", err,tx.(*configTx))
			return err
		}
	default:
		return errors.New("Transaction type not recognized.")
	}

	return nil
}

func (b *Block) addFundsTx(tx *fundsTx) error {

		//I think we don't have to check for nil here as well, since this was already implicitly done with addTx(...)
	if tx.Fee < FEE_MINIMUM {
		err := fmt.Sprintf("Fee (%v) below accepted threshold (%v)\n", tx.Fee, FEE_MINIMUM)
		return errors.New(err)
	}

	//checking if the sender account is already in the local state copy
	if _,exists := b.stateCopy[tx.fromHash]; !exists {
		for _,acc := range State[tx.From] {
			hash := serializeHashContent(acc.Address)
			if hash == tx.fromHash {
				newAcc := Account{}
				newAcc = *acc
				b.stateCopy[tx.fromHash] = &newAcc
			}
		}
	}

	//vice versa for receiver account
	if _,exists := b.stateCopy[tx.toHash]; !exists {
		for _,acc := range State[tx.To] {
			hash := serializeHashContent(acc.Address)
			if hash == tx.toHash {
				newAcc := Account{}
				newAcc = *acc
				b.stateCopy[tx.toHash] = &newAcc
			}
		}
	}

	//rootkey doesn't need to get checked for balance
	//however, txcnt is still increased, makes things a little easiert in the state manipulation
	if !isRootKey(tx.fromHash) {
		if (tx.Amount+tx.Fee) > b.stateCopy[tx.fromHash].Balance {
			return errors.New("Not enough funds to complete the transaction!")
		}
	}

	//check if txcnt makes sense
	if b.stateCopy[tx.fromHash].TxCnt != tx.TxCnt {
		err := fmt.Sprintf("Sender txCnt does not match: %v (tx.txCnt) vs. %v (state txCnt)",tx.TxCnt, b.stateCopy[tx.fromHash].TxCnt)
		return errors.New(err)
	}

	//don't add tx if amount leads to overflow at receiver acc (amount == 0 has already been checked with verify())
	if b.stateCopy[tx.toHash].Balance + tx.Amount > MAX_MONEY {
		err := fmt.Sprintf("Transaction amount (%v) leads to overflow at receiver account balance (%v).\n", tx.Amount, b.stateCopy[tx.toHash].Balance)
		return errors.New(err)
	}

	accSender := b.stateCopy[tx.fromHash]
	accSender.TxCnt += 1
	accSender.Balance -= tx.Amount

	accReceiver := b.stateCopy[tx.toHash]
	accReceiver.Balance += tx.Amount

	b.FundsTxData = append(b.FundsTxData, hashFundsTx(tx))
	writeOpenFundsTx(tx)
	log.Printf("Added tx to the block FundsTxData slice: %v", *tx)
	return nil
}

func (b *Block) addAccTx(tx *accTx) error {

	if tx.Fee < FEE_MINIMUM {
		err := fmt.Sprintf("Fee (%v) below accepted threshold (%v)\n", tx.Fee, FEE_MINIMUM)
		return errors.New(err)
	}

	//at this point the tx has already been verified
	var mapId [8]byte
	accHash := sha3.Sum256(tx.PubKey[:])
	copy(mapId[:],accHash[0:8])
	for _,j := range State[mapId] {
		if bytes.Compare(tx.PubKey[:],j.Address[:]) == 0 {
			return errors.New("Account already exists.")
		}
	}

	b.AccTxData = append(b.AccTxData,hashAccTx(tx))
	writeOpenAccTx(tx)
	log.Printf("Added tx to the AccTxData slice: %v", *tx)
	return nil
}

func (b *Block) addConfigTx(tx *configTx) error {

	if tx.Fee < FEE_MINIMUM {
		err := fmt.Sprintf("Fee (%v) below accepted threshold (%v)\n", tx.Fee, FEE_MINIMUM)
		return errors.New(err)
	}

	b.ConfigTxData = append(b.ConfigTxData,hashConfigTx(tx))
	writeOpenConfigTx(tx)
	log.Printf("Added tx to the ConfigTxData slice: %v", *tx)
	return nil
}

func (b *Block) finalizeBlock() {
	//merkle tree only built from funds transactions
	b.MerkleRoot = buildMerkleTree(b.FundsTxData,b.AccTxData,b.ConfigTxData)
	b.Timestamp = time.Now().Unix()
	copy(b.Beneficiary[:],MinerHash[:])

	//anonymous struct
	partialHash := hashBlock(b)
	nonce := proofOfWork(getDifficulty(), partialHash)
	b.Hash = sha3.Sum256(append(nonce.Bytes(),partialHash[:]...))

	//we need to write the proof at the end of the fixed-size byte array of length 9
	//needs to be decoded by the receiver
	for index,val := range nonce.Bytes() {
		b.Nonce[PROOF_SIZE-len(nonce.Bytes())+index] = val
	}

	//this doesn't need to be hashed, because we already have the merkle tree taking care of consistency
	b.NrFundsTx = uint16(len(b.FundsTxData))
	b.NrAccTx = uint16(len(b.AccTxData))
	b.NrConfigTx = uint8(len(b.ConfigTxData))
}

//this function needs to be split into block syntax/PoW check and actual state change
//because there is the case that we might need to go fetch several blocks in reverse order
//and have to check the blocks first before changing the state in the correct order
func validateBlock(b *Block) error {

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
	for _,block := range blocksToValidate {
		fundsTxs,accTxs,configTxs,err := preValidation(block)
		if err != nil {
			return err
		}
		blockDataMap[block.Hash] = blockData{fundsTxs,accTxs, configTxs,block}
	}

	//no rollback needed, just a new block to validate
	if len(blocksToRollback) == 0 {
		for _,block := range blocksToValidate {
			if err := stateValidation(blockDataMap[block.Hash]); err != nil {
				//if one block fails along the way, we just stop, but this is very unlikely to happen
				return err
			}
			postValidation(blockDataMap[block.Hash])
		}
	} else {
		for _, block := range blocksToRollback {
			if err := validateBlockRollback(block); err != nil {
				return err
			}
		}
		for _, block := range blocksToValidate {
			if err := stateValidation(blockDataMap[block.Hash]); err != nil {
				//if one block fails along the way, we just stop, but this is very unlikely to happen
				return err
			}
			postValidation(blockDataMap[block.Hash])
		}
	}

	return nil
}

func preValidation(b *Block) (fundsTxSlice []*fundsTx, accTxSlice []*accTx, configTxSlice []*configTx, err error) {
	//TODO: make sure none of the transactions are already confirmed
	//check if fundsTxs is syntactically well-formed and signature is correct
	for _, txHash := range b.FundsTxData {
		closeTx := readClosedFundsTx(txHash)
		if closeTx != nil {
			return nil,nil,nil,errors.New("Block validation had fundsTx that was already in a previous block")
		}
		tx := readOpenFundsTx(txHash)
		if tx == nil {
			//TODO: fetch from the network and make sure not in the confirmed map
			return nil,nil,nil,errors.New("FundsTx could not be read.")
		}

		if !(tx).verify() {
			return nil,nil,nil,errors.New("Malformed transaction.")
		}
		fundsTxSlice = append(fundsTxSlice,tx)
	}

	//check if accTxs are syntactically well-formed and signature is correct
	for _, txHash := range b.AccTxData {
		tx := readOpenAccTx(txHash)
		if tx == nil {
			//TODO: fetch from the network and make sure not in the confirmed map
			return nil,nil,nil,errors.New("AccTx could not be read.")
		}
		if !(tx).verify() {
			return nil,nil,nil,errors.New("Malformed transaction.")
		}
		accTxSlice = append(accTxSlice,tx)
	}

	for _, txHash := range b.ConfigTxData {
		tx := readOpenConfigTx(txHash)
		if tx == nil {
			//TODO: fetch from the network and make sure not in the confirmed map
			return nil,nil,nil,errors.New("ConfigTx could not be read.")
		}
		if !(tx).verify() {
			return nil,nil,nil,errors.New("Malformed transaction.")
		}
		configTxSlice = append(configTxSlice,tx)
	}

	startIndex := 0
	for _, singleByte := range b.Nonce {
		if singleByte != 0x00 {
			break
		}
		startIndex++
	}
	nonce := b.Nonce[startIndex:]

	partialHash := hashBlock(b)
	if b.Hash != sha3.Sum256(append(nonce, partialHash[:]...)) || !validateProofOfWork(getDifficulty(), b.Hash) {
		return nil,nil,nil,errors.New("Proof of work is incorrect.")
		log.Println("Proof of work is incorrect.")

	}

	log.Println("Proof of work validation passed.")

	//cmp merkle tree
	if buildMerkleTree(b.FundsTxData, b.AccTxData, b.ConfigTxData) != b.MerkleRoot {
		return nil,nil,nil,errors.New("Merkle Root incorrect.")
		log.Println("Merkle Root incorrect.")
	}

	log.Println("Merkle root hash passed.")
	return fundsTxSlice,accTxSlice,configTxSlice,err
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

	//both collectTxFees as well as collectBlockReward can throw an error when the balance of the miner overflows
	//collect fees for both transaction types
	if err := collectTxFees(data.fundsTxSlice, data.accTxSlice, data.configTxSlice, data.block.Beneficiary); err != nil {
		accStateChangeRollback(data.accTxSlice)
		fundsStateChangeRollback(data.fundsTxSlice)
		return err
	}
	//collect block reward
	if err := collectBlockReward(activeParameters.block_reward, data.block.Beneficiary); err != nil {
		collectTxFeesRollback(data.fundsTxSlice,data.accTxSlice,data.configTxSlice,data.block.Beneficiary)
		accStateChangeRollback(data.accTxSlice)
		fundsStateChangeRollback(data.fundsTxSlice)
		return err
	}

	log.Print("Block validated and state changed accordingly: \n")
	PrintState()

	return nil
}

func postValidation(data blockData) {

	//put all txs from the block from open to close
	for _,tx := range data.fundsTxSlice {
		hash := hashFundsTx(tx)
		writeClosedFundsTx(tx)
		deleteOpenFundsTx(hash)
	}

	for _,tx := range data.accTxSlice {
		hash := hashAccTx(tx)
		writeClosedAccTx(tx)
		deleteOpenAccTx(hash)
	}

	//block consists of system parameter changes
	for _,tx := range data.configTxSlice {
		hash := hashConfigTx(tx)
		writeClosedConfigTx(tx)
		deleteOpenConfigTx(hash)
	}


	//the new system parameters get active if the block was successfully validated
	configStateChange(data.configTxSlice,data.block.Hash)
	writeBlock(data.block)
	collectStatistics(data.block)
}

func hashBlock(b *Block) (hash [32]byte) {

	var buf bytes.Buffer

	blockToHash := struct {
		prevHash    [32]byte
		header     uint8
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

	binary.Write(&buf,binary.BigEndian, blockToHash)
	return sha3.Sum256(buf.Bytes())
}

func encodeBlock(b *Block) (encodedBlock []byte) {

	if b == nil {
		return nil
	}

	//making byte array of all non-byte data
	var timeStamp [8]byte
	var nrFundsTx, nrAccTx [2]byte

	binary.BigEndian.PutUint64(timeStamp[:], uint64(b.Timestamp))
	binary.BigEndian.PutUint16(nrFundsTx[:], b.NrFundsTx)
	binary.BigEndian.PutUint16(nrAccTx[:], b.NrAccTx)

	//reserve space
	encodedBlock = make([]byte,
		BLOCKHEADER_SIZE +
		int(b.NrAccTx) * HASH_LEN +
		int(b.NrFundsTx) * HASH_LEN +
		int(b.NrConfigTx) * HASH_LEN)

	encodedBlock[0] = b.Header

	copy(encodedBlock[1:33],b.Hash[:])
	copy(encodedBlock[33:65],b.PrevHash[:])
	copy(encodedBlock[65:74],b.Nonce[:])
	copy(encodedBlock[74:82],timeStamp[:])
	copy(encodedBlock[82:114],b.MerkleRoot[:])
	copy(encodedBlock[114:146],b.Beneficiary[:])
	copy(encodedBlock[146:148],nrFundsTx[:])
	copy(encodedBlock[148:150],nrAccTx[:])
	encodedBlock[150] = byte(b.NrConfigTx)

	index := BLOCKHEADER_SIZE

	for _,txHash := range b.FundsTxData {
		copy(encodedBlock[index:index+HASH_LEN],txHash[:])
		index += HASH_LEN
	}

	for _,txHash := range b.AccTxData {
		copy(encodedBlock[index:index+HASH_LEN],txHash[:])
		index += HASH_LEN
	}

	for _,txHash := range b.ConfigTxData {
		copy(encodedBlock[index:index+HASH_LEN],txHash[:])
		index += HASH_LEN
	}

	return encodedBlock
}

func decodeBlock(encodedBlock []byte) (b *Block) {

	b = new(Block)

	//time.Now().Unix() return int64, but binary.BigEndian only offers uint64
	var timeStampTmp uint64
	var timeStamp int64
	var nrFundsTx, nrAccTx uint16

	if len(encodedBlock) < BLOCKHEADER_SIZE {
		return nil
	}

	timeStampTmp = binary.BigEndian.Uint64(encodedBlock[74:82])
	nrFundsTx = binary.BigEndian.Uint16(encodedBlock[146:148])
	nrAccTx = binary.BigEndian.Uint16(encodedBlock[148:150])
	timeStamp = int64(timeStampTmp)

	b.Header = encodedBlock[0]
	copy(b.Hash[:],encodedBlock[1:33])
	copy(b.PrevHash[:],encodedBlock[33:65])
	copy(b.Nonce[:],encodedBlock[65:74])
	b.Timestamp = timeStamp
	copy(b.MerkleRoot[:],encodedBlock[82:114])
	copy(b.Beneficiary[:],encodedBlock[114:146])
	b.NrFundsTx = nrFundsTx
	b.NrAccTx = nrAccTx
	b.NrConfigTx = uint8(encodedBlock[150])

	index := BLOCKHEADER_SIZE

	var hash [32]byte
	for cnt := 0; cnt < int(nrFundsTx); cnt++ {
		copy(hash[:],encodedBlock[index:index+HASH_LEN])
		b.FundsTxData = append(b.FundsTxData,hash)
		index += HASH_LEN
	}

	for cnt := 0; cnt < int(nrAccTx); cnt++ {
		copy(hash[:],encodedBlock[index:index+HASH_LEN])
		b.AccTxData = append(b.AccTxData,hash)
		index += HASH_LEN
	}

	for cnt := 0; cnt < int(b.NrConfigTx); cnt++ {
		copy(hash[:],encodedBlock[index:index+HASH_LEN])
		b.ConfigTxData = append(b.ConfigTxData,hash)
		index += HASH_LEN
	}

	return b
}

func (b Block) String() string {
	return fmt.Sprintf("\nHash: %x\n" +
		"Previous Hash: %x\n" +
		"Header: %v\n" +
		"Nonce: %x\n" +
		"Timestamp: %v\n" +
		"MerkleRoot: %x\n" +
		"Beneficiary: %x\n" +
		"Amount of fundsTx: %v\n" +
		"Amount of accTx: %v\n" +
		"Amount of configTx: %v\n",
		b.Hash[0:8],
		b.PrevHash[0:8],
		b.Header,
		b.Nonce,
		b.Timestamp,
		b.MerkleRoot[0:8],
		b.Beneficiary[0:8],
		b.NrFundsTx,
		b.NrAccTx,
		b.NrConfigTx,
	)
}