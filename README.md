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
_Sig_ is the signature (using the private key of the _from_ Account)

- Creating a transaction

The method signature to create a transaction that sends funds from one account to another looks as follows:
```go
constrTx(txCnt uint64, amount uint32, from, to [32]byte, key *ecdsa.PrivateKey) (tx Transaction, err error)
```

### Block
