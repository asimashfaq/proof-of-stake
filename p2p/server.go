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

//The reason we use an additional listener port is because the port the miner connected to this peer
//is not the same as the one it listens to for new connections. When we are queried for neighbors
//we send the IP address in p.conn.RemotAddr() with the listenerPort
type peer struct {
	conn net.Conn
	ch   chan []byte
	l    sync.Mutex
	listenerPort string
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
	go checkHealthService()
	return nil
}

func initiateNewMinerConnection(ipport string) (*peer, error) {

	var conn net.Conn

	conn, err := net.Dial("tcp", ipport)
	p := &peer{conn, nil, sync.Mutex{}, ""}

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
		p := &peer{conn, nil, sync.Mutex{}, ""}
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

