package p2p

import (
	"errors"
	"github.com/lisgie/bazo_miner/protocol"
	"log"
	"net"
	"os"
	"sync"
	"time"
	"encoding/binary"
	"strconv"
)

const (
	MIN_MINERS       = 10
	MAX_MINERS       = 20
	TX_BUFFER        = 10
	BOOTSTRAP_SERVER = "127.0.0.1:8000"
	IPV4ADDR         = 4
	PORT_SIZE        = 2
)

var (
	//List of ip addresses. A connection to a subset of the list will be established as soon as the network health
	//monitor triggers.
	iplistChan chan string

	logger     *log.Logger
	peers      map[*peer]bool
	brdcstMsg  chan []byte
	register   chan *peer
	disconnect chan *peer
	localPort uint16
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
	l    sync.Mutex
}

func Init(port string) error {

	LogFile, _ := os.OpenFile("log/p2p "+time.Now().String(), os.O_RDWR|os.O_CREATE, 0666)
	logger = log.New(LogFile, "", log.LstdFlags)

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

	iplistChan = make(chan string, MIN_MINERS)

	go broadcastService()
	go receiveDataFromMiner()

	//set localPort global, this will be the listening port for incoming connection
	parsedPort, err := strconv.Atoi(port)
	if err != nil {
		return errors.New("Failed to parse port given on the command line")
	}
	localPort = uint16(parsedPort)

	//after this call, there are some peers connected
	// just to test on localhost
	if port != "8000" {
		logger.Print("Start mining as a non-bootstrap node.")
		err := bootstrap()
		if err != nil {
			return err
		}
	} else {
		logger.Print("Start mining as a bootstrap node.")
	}

	go listener(strconv.Itoa(int(localPort)))
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
	go checkHealth()
	return nil
}

func initiateNewMinerConnection(ipport string) (*peer, error) {

	var conn net.Conn

	conn, err := net.Dial("tcp", ipport)
	p := &peer{conn, nil, sync.Mutex{}}

	if err != nil {
		return nil, err
	}

	//We need to additionally send either IP:Port or Port in order to construct a valid first message
	var port [2]byte
	binary.BigEndian.PutUint16(port[:], localPort)
	packet := BuildPacket(MINER_PING, port[:])
	conn.Write(packet)
	header, _, err := rcvData(p)
	if err != nil || header.TypeID != MINER_PONG {
		return nil, errors.New("Connecting to bootstrap server failed.")
	}

	return p, nil
}

func listener(localPort string) {

	listener, err := net.Listen("tcp", ":"+localPort)
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
		p := &peer{conn, nil, sync.Mutex{}}
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
		processNeighborRes(p, payload)
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

//single goroutine that makes sure the system is well connected
func checkHealth() {

	for {
		time.Sleep(time.Minute)
		if len(peers) >= MIN_MINERS {
			continue
		}

		select {
		case ipaddr := <-iplistChan:
			logger.Printf("New IP rcvd through channel: %v\n", ipaddr)
			p, err := initiateNewMinerConnection(ipaddr)
			if err != nil {
				logger.Printf("Failed to initiate connection with IP address: %v\n", ipaddr)
				continue
			}
			go minerConn(p)
		default:
			neighborReq()
			logger.Print("30 seconds passed and no address received.\n")
		}
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
		sendData(p, msg)
	}
}
