package bc

type Account struct {
	Nonce, Balance int64
	Id [64]byte
}

