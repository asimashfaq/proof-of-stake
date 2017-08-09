package p2p

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
)

var (
	//List of ip addresses. A connection to a subset of the list will be established as soon as the network health
	//monitor triggers.
	localConn  string
	peers      peersStruct

	logger     *log.Logger

	iplistChan = make(chan string, MIN_MINERS)
	brdcstMsg = make(chan []byte)
	register = make(chan *peer)
	disconnect = make(chan *peer)
)

//Entry point for p2p package
func Init(connTuple string) error {

	logInit()

	//Initialize peer map
	peers.peerConns = make(map[*peer]bool)

	//Start all services that are running concurrently
	go broadcastService()
	go checkHealthService()
	go timeService()
	go receiveBlockFromMiner()

	//Set localPort global, this will be the listening port for incoming connection
	localConn = connTuple
	ipport := strings.Split(localConn, ":")
	if ipport[1] != "8000" {
		err := bootstrap()
		if err != nil {
			return err
		}
	}

	//Listen for all subsequent incoming connections on specified local address/listening port
	go listener(localConn)
	return nil
}

func bootstrap() error {
	//Connect to bootstrap server. To make it more fault-tolerant, we can increase the number of bootstrap servers in
	//the future. initiateNewMinerConn(...) starts with MINER_PING to perform the initial handshake message
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

	//Check if we already established a connection with that ip or if the ip belongs to us
	if peerExists(ipport) {
		return nil, errors.New(fmt.Sprintf("Connection with %v already established.\n", ipport))
	}

	if peerSelfConn(ipport) {
		return nil, errors.New(fmt.Sprintf("Cannot self-connect %v.\n", ipport))
	}

	//Open up a tcp connection and instantiate a peer struct, wait for adding it to the peerStruct before we finalize
	//the handshake
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

	//Wait for the other party to finish the handshake with the corresponding message
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
	//Extracts the port from our localConn variable (which is in the form IP:Port)
	localPort, err := strconv.Atoi(strings.Split(localConn, ":")[1])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Parsing port failed: %v\n", err))
	}
	binary.BigEndian.PutUint16(portBuf[:], uint16(localPort))
	packet := BuildPacket(MINER_PING, portBuf)

	return packet, nil
}

func listener(ipport string) {

	//Listen on all interfaces, this NAT stuff easier
	listener, err := net.Listen("tcp", ":"+strings.Split(ipport,":")[1])
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
	//Give the peer a channel
	p.ch = make(chan []byte)
	//Register withe the broadcast service and start the additional writer
	register <- p
	go peerBroadcast(p)

	for {
		header, payload, err := rcvData(p)
		if err != nil {
			logger.Printf("Miner disconnected: %v\n", err)
			//In case of a comm fail, disconnect cleanly from the broadcast service
			disconnect <- p
			return
		}

		processIncomingMsg(p, header, payload)
	}
}
