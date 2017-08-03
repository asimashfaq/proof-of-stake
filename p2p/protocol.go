package p2p

import "fmt"

const HEADER_LEN = 5

//Mapping constants
const (
	FUNDSTX_BRDCST  = 1
	ACCTX_BRDCST    = 2
	CONFIGTX_BRDCST = 3
	BLOCK_BRDCST    = 4

	FUNDSTX_REQ  = 10
	ACCTX_REQ    = 11
	CONFIGTX_REQ = 12
	BLOCK_REQ    = 13
	ACC_REQ      = 14

	FUNDSTX_RES  = 20
	ACCTX_RES    = 21
	CONFIGTX_RES = 22
	BLOCK_RES    = 23
	ACC_RES      = 24

	NEIGHBOR_REQ = 30

	NEIGHBOR_RES = 40

	TIME_BRDCST = 50

	MINER_PING = 100
	MINER_PONG = 101

	//Used to signal error
	NOT_FOUND = 110
)

type Header struct {
	Len    uint32
	TypeID uint8
}

func (header Header) String() string {
	return fmt.Sprintf(
		"Length: %v\n"+
			"TypeID: %v\n",
		header.Len,
		header.TypeID,
	)
}