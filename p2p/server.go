package p2p

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/lisgie/bazo_miner/protocol"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
)

var (
	//List of ip addresses. A connection to a subset of the list will be established as soon as the network health
	//monitor triggers.
	iplistChan chan string
	peers      peersStruct
	logger     *log.Logger
	brdcstMsg  chan []byte
	register   chan *peer
	disconnect chan *peer
	localConn  string
)

//we need to decode incoming transactions, therefore type is needed
//for outgoing transactions, the p2p package needs the information to build the proper header
type TxInfo struct {
	TxType  uint8
	Payload []byte
}

func Init(connTuple string) error {

	logInit()

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
	go timeService()
	go receiveDataFromMiner()

	//set localPort global, this will be the listening port for incoming connection
	localConn = connTuple
	ipport := strings.Split(localConn, ":")
	if ipport[1] != "8000" {
		logger.Print("Start mining as a non-bootstrap node\n")
		err := bootstrap()
		if err != nil {
			return err
		}
	} else {
		logger.Print("Start mining as a bootstrap node\n")
	}

	go listener(localConn)
	return nil
}

func bootstrap() error {
	//connect to bootstrap server
	//initiate MINER_PING
	//add to connection list
	p, err := initiateNewMinerConnection(BOOTSTRAP_SERVER)
	if err != nil {
		logger.Printf("Initiating new miner connection failed: %v\n", err)
		return err
	}

	go minerConn(p)
	return nil
}

func initiateNewMinerConnection(ipport string) (*peer, error) {

	var conn net.Conn

	//check if we already established a connection with that ip or if the ip belongs to us
	if peerExists(ipport) {
		return nil, errors.New(fmt.Sprintf("Connection with %v already established.\n", ipport))
	}

	if peerSelfConn(ipport) {
		return nil, errors.New(fmt.Sprintf("Cannot self-connect %v.\n", ipport))
	}

	conn, err := net.Dial("tcp", ipport)
	p := &peer{conn, nil, sync.Mutex{}, strings.Split(ipport, ":")[1], 0}

	if err != nil {
		return nil, err
	}

	packet, err := prepareHandshake()
	if err != nil {
		return nil, err
	}

	conn.Write(packet)
	header, _, err := rcvData(p)
	if err != nil || header.TypeID != MINER_PONG {
		return nil, errors.New(fmt.Sprintf("Failed to complete miner handshake: %v\n", err))
	}

	return p, nil
}

func prepareHandshake() ([]byte, error) {
	//We need to additionally send our local listening port in order to construct a valid first message
	//This will be the only time we need it so we don't save it
	portBuf := make([]byte, PORT_SIZE)
	localPort, err := strconv.Atoi(strings.Split(localConn, ":")[1])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Parsing port failed: %v\n", err))
	}
	binary.BigEndian.PutUint16(portBuf[:], uint16(localPort))
	packet := BuildPacket(MINER_PING, portBuf)

	return packet, nil
}

func listener(ipport string) {

	listener, err := net.Listen("tcp", ipport)
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
		p := &peer{conn, nil, sync.Mutex{}, "", 0}
		go handleNewConn(p)
	}
}

func handleNewConn(p *peer) {

	logger.Printf("New incoming connection: %v\n", p.conn.RemoteAddr().String())
	header, payload, err := rcvData(p)
	if err != nil {
		logger.Printf("Failed to handle incoming connection: %v\n", err)
		return
	}

	processIncomingMsg(p, header, payload)
}

func minerConn(p *peer) {

	logger.Printf("Adding a new miner: %v\n", p.getIPPort())

	ch := make(chan []byte)
	//give the peer a channel
	p.ch = ch
	register <- p
	go peerBroadcast(p)

	for {
		header, payload, err := rcvData(p)
		if err != nil {
			logger.Printf("Miner disconnected: %v\n", err)
			disconnect <- p
			return
		}

		processIncomingMsg(p, header, payload)
	}
}
