package protocol

type Transaction interface {
	Hash() [32]byte
	//Encode() []byte
	//Decode([]byte) Transaction
	//GetFee() uint64
}
