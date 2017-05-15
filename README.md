# Technical Description

## Buildig Blocks

### Account

Every participant needs to have an account. The Account struct looks as follows:
```go
type Account struct {
	Address [64]byte
	TxCnt, Balance uint64
}
```
P-256 (see FIPS 186-3, section D.2.3) is used as elliptic curve. The _Address_ is the concatenation of two coordinates and forms the public key of an account. The _TxCnt_ is increased for every transaction (start value: 0), this is needed to prevent replay attacks. _Balance_ is how much coins are attached to the Account.

### State

The cryptocurrency is account-based (compared to unspent transaction-based, cf. Bitcoin), the state of all accounts are maintained and stored in a hashmap:
```go
var State map[[32]byte]Account
```
The key _[32]byte_ is the SHA3-256 hash of the account's public key. State manipulation is only done in the file _state.go_. The state is updated when funds are transferred (or rolled back) and if new accounts are added.

### Transaction

Because the state maintains a mapping from key to hash(key), transactions use the hash as an account specifier which saves 64 Bytes per transaction (compared to sending full public keys with every transaction).

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
	Sig [64]byte
	Payload txPayload
}

type txPayload struct {
	Nonce uint64
	Amount uint32
	From, To [32]byte
}
```
A _fundsTx_ consists of a signature (_Sig_) and the relevant transaction data: _Nonce_ is the _TxCnt_ from the sender account. _Amount_ is a 32-bit number (might be changed to 64-bit later), speicifying the amount of money and _From_, _To_ are the hashes of the participating accounts. _Sig_ is the signature (using the private key of the _from_ Account) of a Sha3-256 hash of the _TxInfo_. The total size of a transaction is fixed at 140 Bytes (or 144 Bytes if we allow 64-bit _Amount_).

The method signature to create a _fundsTx_ transaction that sends funds from one account to another looks as follows:
```go
constrTx(txCnt uint64, amount uint32, from, to [32]byte, key *ecdsa.PrivateKey) (tx Transaction, err error)
```
The _accTx_ is the transaction type that adds a new account to the system:
```go
type accTx struct {
	Sig [64]byte
	PubKey [64]byte
}
```
_PubKey_ is the public key of the new account that has been generated. In a first version of the system, the validity of a new account is checked against the company's public key (company signs with pubkey and broadcast ths tx to the network). And only if the signature matches, the account is included into the state.

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
	Version uint8
	Proof [ProofSize]byte
	Timestamp int64
	Difficulty uint8
	MerkleRoot [32]byte
	stateCopy map[[32]byte]Account
	FundsTxData []fundsTx
	AccTxData []accTx
}
```
_Hash_ is the global identifier of the block, _prevHash_ is the hash of the previous block. _Version_ is by default set to 1, this allows to make protocol changes later on. _Proof_ is a fixed byte array of size _ProofSize_, acting as a PoW (Proof of Work). This byte array, appended by the hash of several other fields and hashed again has to fulfill the properties that the first _Difficulty_ of bits of the resulting hash (which is equal the _Hash_ field) will be 0. _ProofSize_ is a constant set to 9, which, based on some calculations, is a good trade-off between memory and future network hash rate (e.g., even at Bitcoin's network hash rate, 9 Bytes is enough). _Merkleroot_ is the hash of the merkle tree consisting of all transactions. This will be used by light clients to verify if certain transactions took place by querying full nodes for the relevant merkle path. _FundsTxData_ is a slice, consisting of all _fundsTx_ transactions within this block (vice versa with _AccTxdata_).

- Adding transactions

Received transactions are verified and added to the current block if all checks pass.

- Block Validation

Block validation consists of the following:

	* Checking if prevHash makes sense
	* Checking if correct proof of work depending on the level of difficulty
	* Recalculate merkle root and check if identical
	* Check if all transactions are well-formed

If all checks were successful, the state is updated by going through each transaction sequentially. If there is an illegal transaction (e.g., sending money from an account not posessing the needed funds), all changes are reverted and "rolled back" to the state before the block validation.
