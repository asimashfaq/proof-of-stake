# Technical Description

## Buildig Blocks

### Account

Every participant needs to have an account. The Account struct looks as follows:
```go
type Account struct {
	Hash [32]byte
	Address [64]byte
	Balance uint64
	TxCnt uint32
}
```
P-256 (see FIPS 186-3, section D.2.3) is used as elliptic curve. The _Address_ is the concatenation of two coordinates and forms the public key of an account. The _TxCnt_ is increased for every transaction (start value: 0), this is needed to prevent replay attacks. _Balance_ is how much coins are attached to the Account.

### State

The cryptocurrency is account-based (compared to unspent transaction-based, cf. Bitcoin), the state of all accounts are maintained and stored in a hashmap:
```go
var State map[[8]byte][]*Account
```
The key _[8]byte_ is the first 8 bytes of the SHA3-256 hash of the account's public key. The reason this is done is to reduce funds transactions to below 100 bytes. The value of the map is a slice of _Account_ pointers (several accounts might have identical first 8 bytes in their hash). State manipulation is only done in the file _state.go_. The state is updated when funds are transferred (or rolled back) and if new accounts are added.

### Transaction

- Transaction Types

Two transaction types exist, namely _fundsTx_ (moving funds around) and _accTx_ (creating new addresses), both implementing the _transaction_ interface (not really sure about that yet, but I think it simplifies encoding and saves some code): 
```go
type transaction interface {
	verify() bool
}
```
The _fundsTx_ structure is as follows:
```go
type fundsTx struct {
	Header byte
	Amount [4]byte
	Fee [2]byte
	TxCnt [3]byte
	From [8]byte
	fromHash [32]byte
	To [8]byte
	toHash [32]byte
	Xored [24]byte
	Sig [40]byte
}
```
A _fundsTx_ consists of a signature (_Sig_) and the relevant transaction data: _TxCnt_ is the _TxCnt_ from the sender account. _Amount_ is a 32-bit number (might be changed to 64-bit later), speicifying the amount of money and _From_, _To_ are the first 8 bytes of the hashes of the participating accounts (_fromHash_, _toHash_ are the full hashes of the accounts. However, they won't be exported when serializing). _Sig_ is the signature (using the private key of the _from_ Account). _Xored_ is the last 24 bytes of _fromHash_, _toHash_ and the first 24 bytes of the signature xored (compressing data at the cost of potential conflicts). This way the tx size is < 100 bytes. Not really sure yet about the size of _Amount_ and _Fee_.

The other transaction type is to create a new account:
```go
type accTx struct {
	Issuer [32]byte
	Fee [2]byte
	Sig [64]byte
	PubKey [64]byte
}
```
_PubKey_ is the public key of the new account that has been generated. In a first version of the system, the validity of a new account is checked against the _Issuer_ public key (some entity signs with private key and broadcast the tx to the network). And only if the signature matches, the account is included into the state. _Fee_ is needed to incentivize miners to include this transaction.

- Verifying a transaction
The signature to verify a transaction is the following
```go
(tx Transaction) verifyTx() bool
```
Both transaction types (_fundsTx_ and _accTx_) implement the _verifyTx()_ method and check whether the transaction is syntactically well-formed and makes sense in the state context (e.g., does the sender have enough funds, does the sender and receiver exist etc.)

### Block
The structure of a block is as follows:
```go
type Block struct {
	Hash [32]byte
	PrevHash [32]byte
	Version uint8 //future updates
	Proof [ProofSize]byte //ProofSize set to 72-bit, enough even if the network gets really large
	Timestamp int64
	Difficulty uint8
	MerkleRoot [32]byte
	Beneficiary [32]byte
	//this field will not be exported, this is just to avoid race conditions with the global state
	stateCopy map[[32]byte]*Account
	FundsTxData []fundsTx
	AccTxData []accTx
}```
_Hash_ is the global identifier of the block, _prevHash_ is the hash of the previous block. _Version_ is by default set to 1, this allows to make protocol changes later on. Every central element, which are planned to be removed further, are bound to the version number. _Proof_ is a fixed byte array of size _ProofSize_, acting as a PoW (Proof of Work). This byte array, appended by the hash of several other fields and hashed again has to fulfill the properties that the first _Difficulty_ of bits of the resulting hash (which is equal the _Hash_ field) will be 0. _ProofSize_ is a constant set to 9, which, based on some calculations, is a good trade-off between memory and future network hash rate (e.g., even at Bitcoin's network hash rate, 9 Bytes is enough). _Merkleroot_ is the hash of the merkle tree consisting of all transactions. This will be used by light clients to verify if certain transactions took place by querying full nodes for the relevant merkle path. _FundsTxData_ is a slice, consisting of all _fundsTx_ transactions within this block (vice versa with _AccTxdata_).

- Adding transactions

Received transactions are verified and added to the current block if all checks pass.

- Block Validation

Block validation consists of the following:

	* Checking if prevHash makes sense
	* Checking if correct proof of work depending on the level of difficulty
	* Recalculate merkle root and check if identical
	* Check if all transactions are well-formed

If all checks were successful, the state is updated by going through each transaction sequentially. If there is an illegal transaction (e.g., sending money from an account not posessing the needed funds), all changes are reverted and "rolled back" to the state before the block validation.
