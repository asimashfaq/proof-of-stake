package bc

type Account struct {
	Address [64]byte
	TxCnt, Balance uint64
}