package p2p

import (
	"errors"
	"log"
	"net"
	"os"
	"strconv"
	"time"
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
	peers      map[peer]bool
	brdcstMsg  chan []byte
	register   chan peer
	disconnect chan peer

	TxsIn   chan TxInfo
	BlockIn chan []byte

	TxsOut   chan TxInfo
	BlockOut chan []byte
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
}

//4 channels for communication with miner, blocks in/out and txs in/out
func Init(port string) error {

	LogFile, _ := os.OpenFile("log/p2p "+time.Now().String(), os.O_RDWR|os.O_CREATE, 0666)
	logger = log.New(LogFile,"",log.LstdFlags)


	TxsIn = make(chan TxInfo, TX_BUFFER)
	BlockIn = make(chan []byte)
	TxsOut = make(chan TxInfo, TX_BUFFER)
	BlockOut = make(chan []byte)

	peers = make(map[peer]bool)
	brdcstMsg = make(chan []byte)
	register = make(chan peer)
	disconnect = make(chan peer)

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
	conn, err := initiateNewMinerConnection(BOOTSTRAP_SERVER)
	if err != nil {
		logger.Printf("%v\n", err)
		return err
	}
	go minerConn(conn)

	//once connected to the bootstrap, get his neighbors as well
	packet := BuildPacket(NEIGHBOR_REQ, nil)
	conn.Write(packet)

	iplist := neighborReq()
	if iplist == nil {
		return nil
	}

	//parse the incoming ipv4 addresses

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
	}

	return nil
}

func initiateNewMinerConnection(ip string) (net.Conn, error) {

	var conn net.Conn

	conn, err := net.Dial("tcp", ip+":"+strconv.Itoa(PORT))

	if err != nil {
		return nil, err
	}

	packet := BuildPacket(MINER_PING, nil)
	conn.Write(packet)

	header, _, err := rcvData(conn)
	if err != nil || header.TypeID != MINER_PONG {
		return nil, errors.New("Connecting to bootstrap server failed.")
	}

	return conn, nil
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
		go handleNewConn(conn)
	}
}

func handleNewConn(conn net.Conn) {

	logger.Printf("New incoming connection: %v\n", conn.RemoteAddr().String())
	header, payload, err := rcvData(conn)
	if err != nil {
		logger.Printf("%v\n", err)
	}

	processRequest(conn, header, payload)

	if header.TypeID == MINER_PING {
		go minerConn(conn)
	} else {
		conn.Close()
	}
}

func processRequest(conn net.Conn, header *Header, payload []byte) {

	logger.Printf("Received request from %v with following header:\n%v", conn.RemoteAddr().String(), header)
	switch header.TypeID {
	case FUNDSTX_BRDCST:
		forwardTxToMiner(conn, payload, FUNDSTX_BRDCST)
	case ACCTX_BRDCST:
		forwardTxToMiner(conn, payload, ACCTX_BRDCST)
	case CONFIGTX_BRDCST:
		forwardTxToMiner(conn, payload, CONFIGTX_BRDCST)
	case BLOCK_BRDCST:
		forwardBlockToMiner(conn, payload)
	case FUNDSTX_REQ:
		txRes(conn, payload, FUNDSTX_REQ)
	case ACCTX_REQ:
		txRes(conn, payload, ACCTX_REQ)
	case CONFIGTX_REQ:
		txRes(conn, payload, CONFIGTX_REQ)
	case BLOCK_REQ:
		blockRes(conn, payload)
	case ACC_REQ:
		accRes(conn, payload)
	case MINER_PING:
		pongRes(conn, payload)
	case NEIGHBOR_REQ:
		neighborRes(conn, payload)
	}
}

func receiveDataFromMiner() {

	for {
		select {
		case block := <-BlockOut:
			logger.Printf("Received a block from the miner for broadcasting.")
			toBrdcst := BuildPacket(BLOCK_BRDCST, block)
			brdcstMsg <- toBrdcst
		case txInfo := <-TxsOut:
			logger.Printf("Received a transaction from the miner for broadcasting: ID: %v.\n", txInfo.TxType)
			toBrdcst := BuildPacket(txInfo.TxType, txInfo.Payload)
			brdcstMsg <- toBrdcst
		}
	}
}

//we can't broadcast incoming messages directly, need to forward them to the miner (to check if
//the tx has already been broadcast before, whether it was a valid tx at all)
func forwardTxToMiner(conn net.Conn, payload []byte, brdcstType uint8) {
	logger.Printf("Received a transaction (ID: %v) from %v.\n", brdcstType, conn.RemoteAddr().String())
	TxsIn <- TxInfo{brdcstType, payload}
}
func forwardBlockToMiner(conn net.Conn, payload []byte) {
	logger.Printf("Received a block from %v.\n", conn.RemoteAddr().String())
	BlockIn <- payload
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
			p.conn.Close()
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

func minerConn(conn net.Conn) {

	logger.Printf("Adding a new miner: %v\n", conn.RemoteAddr().String())

	ch := make(chan []byte)
	p := peer{conn, ch}
	register <- p
	go peerWriter(p)

	for {
		header, payload, err := rcvData(p.conn)
		if err != nil {
			disconnect <- p
		}

		processRequest(conn, header, payload)
	}
}

//will be our broadcast mechanism
func peerWriter(p peer) {
	for msg := range p.ch {
		logger.Printf("Sending payload to %v\n", p.conn.RemoteAddr().String())
		p.conn.Write(msg)
	}
}
