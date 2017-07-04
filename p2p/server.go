package p2p

import (
	"errors"
	"log"
	"net"
	"os"
	"strconv"
	"time"
	"fmt"
	"sync"
	"github.com/lisgie/bazo_miner/protocol"
)

const (
	PORT             = 8000
	MIN_MINERS       = 10
	MAX_MINERS       = 20
	TX_BUFFER        = 10
	BOOTSTRAP_SERVER = "127.0.0.1"
	IPV4ADDR         = 4
)

var (
	//LogFile    *os.File
	logger *log.Logger
	peers      map[*peer]bool
	brdcstMsg  chan []byte
	register   chan *peer
	disconnect chan *peer
)

//we need to decode incoming transactions, therefore type is needed
//for outgoing transactions, the p2p package needs the information to build the proper header
type TxInfo struct {
	TxType  uint8
	Payload []byte
}

type peer struct {
	conn net.Conn
	ch   chan []byte
	l sync.Mutex
}

//4 channels for communication with miner, blocks in/out and txs in/out
func Init(port string) error {

	LogFile, _ := os.OpenFile("log/p2p "+time.Now().String(), os.O_RDWR|os.O_CREATE, 0666)
	logger = log.New(LogFile,"",log.LstdFlags)


	TxsIn = make(chan TxInfo, TX_BUFFER)
	BlockIn = make(chan []byte)
	TxsOut = make(chan TxInfo, TX_BUFFER)
	BlockOut = make(chan []byte)

	//channels for specific miner requests
	BlockReqChan = make(chan []byte)
	FundsTxChan = make(chan *protocol.FundsTx)
	AccTxChan = make(chan *protocol.AccTx)
	ConfigTxChan = make(chan *protocol.ConfigTx)

	peers = make(map[*peer]bool)
	brdcstMsg = make(chan []byte)
	register = make(chan *peer)
	disconnect = make(chan *peer)

	go broadcastService()
	go receiveDataFromMiner()
	//go checkHealth()

	//after this call, there are some peers connected

	//just to test on localhost
	if port != "8000" {
		logger.Print("Start mining as a non-bootstrap node.")
		err := bootstrap()
		if err != nil {
			return err
		}
	} else {
		logger.Print("Start mining as a bootstrap node.")
	}

	go listener(port)
	return nil
}

func bootstrap() error {
	//connect to bootstrap server
	//initiate MINER_PING
	//add to connection list
	p, err := initiateNewMinerConnection(BOOTSTRAP_SERVER)
	if err != nil {
		logger.Printf("%v\n", err)
		return err
	}
	go minerConn(p)

	//once connected to the bootstrap, get his neighbors as well

	/*fmt.Printf("#%v\n", iplist)
	if iplist == nil {
		return nil
	}

	//parse the incoming ipv4 addresses
	fmt.Printf("%v\n", iplist)
	index := 0
	for cnt := 0; cnt < len(iplist)/IPV4ADDR; cnt++ {
		var addr string
		for singleAddr := 0; singleAddr < IPV4ADDR; singleAddr++ {
			tmp := int(iplist[singleAddr])
			addr += strconv.Itoa(tmp)
			addr += "."
		}
		//remove the trailing dot
		conn, err := initiateNewMinerConnection(addr[0 : len(addr)-1])
		if err != nil {
			logger.Printf("Connection to miner addr %v could not be established.\n", addr[0:len(addr)-1])
			continue
		}
		go minerConn(conn)
		index += IPV4ADDR
	}*/

	return nil
}

func initiateNewMinerConnection(ip string) (*peer, error) {

	var conn net.Conn

	conn, err := net.Dial("tcp", ip+":"+strconv.Itoa(PORT))
	p := &peer{conn,nil,sync.Mutex{}}

	if err != nil {
		return nil, err
	}

	packet := BuildPacket(MINER_PING, nil)
	conn.Write(packet)

	header, _, err := rcvData(p)
	if err != nil || header.TypeID != MINER_PONG {
		return nil, errors.New("Connecting to bootstrap server failed.")
	}

	return p, nil
}

func listener(port string) {

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Printf("%v\n", err)
		return
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Printf("%v\n", err)
			continue
		}
		p := &peer{conn,nil,sync.Mutex{}}
		go handleNewConn(p)
	}
}

func handleNewConn(p *peer) {

	logger.Printf("New incoming connection: %v\n", p.conn.RemoteAddr().String())
	header, payload, err := rcvData(p)
	if err != nil {
		logger.Printf("%v\n", err)
		return
	}

	processIncomingMsg(p, header, payload)

	if header.TypeID == MINER_PING {
		go minerConn(p)
	} else {
		p.conn.Close()
	}
}

func processIncomingMsg(p *peer, header *Header, payload []byte) {

	logger.Printf("Received request from %v with following header:\n%v", p.conn.RemoteAddr().String(), header)
	switch header.TypeID {
		//BROADCASTING
	case FUNDSTX_BRDCST:
		forwardTxToMiner(p, payload, FUNDSTX_BRDCST)
	case ACCTX_BRDCST:
		forwardTxToMiner(p, payload, ACCTX_BRDCST)
	case CONFIGTX_BRDCST:
		forwardTxToMiner(p, payload, CONFIGTX_BRDCST)
	case BLOCK_BRDCST:
		forwardBlockToMiner(p, payload)

		//Miner Requests
	case FUNDSTX_REQ:
		txRes(p, payload, FUNDSTX_REQ)
	case ACCTX_REQ:
		txRes(p, payload, ACCTX_REQ)
	case CONFIGTX_REQ:
		txRes(p, payload, CONFIGTX_REQ)
	case BLOCK_REQ:
		blockRes(p, payload)
	case ACC_REQ:
		accRes(p, payload)
	case MINER_PING:
		pongRes(p, payload)
	case NEIGHBOR_REQ:
		neighborRes(p, payload)

		//Miner Responses
	case NEIGHBOR_RES:
		fmt.Printf("%v\n%v\n%v\n", p, header, payload)
	case BLOCK_RES:
		forwardBlockReqToMiner(p, payload)
	case FUNDSTX_RES:
		forwardTxReqToMiner(p, payload, FUNDSTX_RES)
	case ACCTX_RES:
		forwardTxReqToMiner(p, payload, ACCTX_RES)
	case CONFIGTX_RES:
		forwardTxReqToMiner(p, payload, CONFIGTX_RES)
	}
}

//this is not accessed concurrently, one single goroutine
func broadcastService() {
	logger.Print("Start broadcasting service.")
	for {
		select {
		//broadcasting all messages
		case msg := <-brdcstMsg:
			for p := range peers {
				p.ch <- msg
			}
		case p := <-register:
			peers[p] = true
		case p := <-disconnect:
			delete(peers, p)
			close(p.ch)
		}
	}
}

//single goroutine that makes sure that system is well connected
func checkHealth() {

	for {
		if len(peers) >= MIN_MINERS {
			time.Sleep(10 * time.Second)
			continue
		}
		//initiate new connection if not enough
		//and call go outgoingConn(conn)
	}
}

func minerConn(p *peer) {

	logger.Printf("Adding a new miner: %v\n", p.conn.RemoteAddr().String())

	ch := make(chan []byte)
	//give the peer a channel
	p.ch = ch
	register <- p
	go peerWriter(p)

	for {
		header, payload, err := rcvData(p)
		if err != nil {
			logger.Printf("%v\n", err)
			disconnect <- p
			return
		}

		processIncomingMsg(p, header, payload)
	}
}

//will be our broadcast mechanism
func peerWriter(p *peer) {
	for msg := range p.ch {
		logger.Printf("Sending payload to %v\n", p.conn.RemoteAddr().String())
		sendData(p,msg)
	}
}
