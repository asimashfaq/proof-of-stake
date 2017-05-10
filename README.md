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
The key _[32]byte_ is the SHA3-256 hash of the account's public key.

### Transaction

Because the state maintains a mapping from key to hash(key), transactions use the hash as an account specifier which saves 64 Bytes per transaction (compared to sending full public keys with every transaction). The transaction structure is as follows:
```go
type Transaction struct {
	Sig [64]byte
	Info TxInfo
}

type TxInfo struct {
	Nonce uint64
	Amount uint32
	From, To [32]byte
}
```
A transaction consists of a signature (_Sig_) and the relevant transaction data: _Nonce_ is the _TxCnt_ from the sender account. _Amount_ is a 32-bit number (might be changed to 64-bit later), speicifying the amount of money and _From_, _To_ are the hashes of the participating accounts. _Sig_ is the signature (using the private key of the _from_ Account) of a Sha3-256 hash of the _TxInfo_. The total size of a transaction is fixed at 140 Bytes.

- Creating a transaction
The method signature to create a transaction that sends funds from one account to another looks as follows:
```go
constrTx(txCnt uint64, amount uint32, from, to [32]byte, key *ecdsa.PrivateKey) (tx Transaction, err error)
```

- Verifying a transaction
The signature to verify a transaction is the following
```go
(tx Transaction) verifyTx() bool
```
Checks whether the signature matches the public key of the sender (proof that the sender was in posession of the corresponding private key).

- Transaction Types (tbd)

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
	TxData []Transaction
}
```
_Hash_ is the global identifier of the block, _prevHash_ is the hash of the previous block. _Version_ is by default set to 1, this allows to make protocol changes later on. _Proof_ is a fixed byte array of size _ProofSize_, acting as a PoW (Proof of Work). This byte array, appended by the hash of several other fields and hashed again has to fulfill the properties that the first _Difficulty_ of bits of the resulting hash (which is equal the _Hash_ field) will be 0. _ProofSize_ is a constant set to 9, which, based on some calculations, is a good trade-off between memory and future network hash rate (e.g., even at Bitcoin's network hash rate, 9 Bytes is enough). _Merkleroot_ is the hash of the merkle tree consisting of all transactions. This will be used by light clients to verify if certain transactions took place by querying full nodes for the relevant merkle path. _TxData_ is a slice, consisting of all transactions within this block.

- Add transaction to block

Check if well-formed transaction and legal in terms of state change

- Finalize block

Calculate merkle tree and proof of work, before broadcasting to the network

- Validate block

Block validation consists of the following:

	* Checking if prevHash makes sense
	* Checking if correct proof of work
	* Recalculate merkle root and check if identical
	* Check if all transactions are well-formed

If all checks were successful, the state is updated by going through each transaction sequentially. If there is an illegal transaction (e.g., sending money from an account not posessing the needed funds), all changes are reverted and "rolled back" to the state before the block validation.
