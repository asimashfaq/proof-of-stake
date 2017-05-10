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

The cryptocurrency is account-based (compared to unspent transaction-based, cf. Bitcoin), the state is saved in 

### Transaction

### Block
