package storage

const (
	ACC_SIZE = 108
	FUNDSTX_SIZE = 100
	ACCTX_SIZE = 169
)

//not accessible from outside
var state map[[8]byte][ACC_SIZE]byte
var rootAccs map[[32]byte][ACC_SIZE]byte
var blocks map[[32]byte][]byte
var opentxs map[[32]byte][]byte
var closedtxs map[[32]byte][]byte

func Init() {

	state = make(map[[8]byte][ACC_SIZE]byte)
	rootAccs = make(map[[32]byte][ACC_SIZE]byte)
	blocks = make(map[[32]byte][]byte)
	opentxs = make(map[[32]byte][]byte)
	closedtxs = make(map[[32]byte][]byte)
}

