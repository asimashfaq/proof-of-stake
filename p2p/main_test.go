package p2p

import (
	"github.com/lisgie/bazo_miner/protocol"
	"os"
	"testing"
)

var (
	MINER_IPPORT = "127.0.0.1:8000"
)

func TestMain(m *testing.M) {

	logInit()
	localConn = "127.0.0.1:9000"

	TxsIn = make(chan TxInfo, TX_BUFFER)
	BlockIn = make(chan []byte)
	TxsOut = make(chan TxInfo, TX_BUFFER)
	BlockOut = make(chan []byte)

	//channels for specific miner requests
	BlockReqChan = make(chan []byte)
	FundsTxChan = make(chan *protocol.FundsTx)
	AccTxChan = make(chan *protocol.AccTx)
	ConfigTxChan = make(chan *protocol.ConfigTx)

	peers.peerConns = make(map[*peer]bool)
	brdcstMsg = make(chan []byte)
	register = make(chan *peer)
	disconnect = make(chan *peer)

	iplistChan = make(chan string, MIN_MINERS)

	go broadcastService()
	go checkHealthService()
	go receiveDataFromMiner()

	//bootstrap server
	go listener("127.0.0.1:8000")

	os.Exit(m.Run())
}
