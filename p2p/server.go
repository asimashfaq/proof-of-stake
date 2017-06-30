package p2p

import (
	"bufio"
	"log"
	"net"
	"os"
	"time"
	"fmt"
	"errors"
)

const (
	PORT       = 8000
	MIN_MINERS = 10
	MAX_MINERS = 20
	TX_BUFFER  = 10
	BOOTSTRAP_SERVER = "127.0.0.1:8000"
)

var (
	LogFile    *os.File
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

type peer chan<- []byte

//4 channels for communication with miner, blocks in/out and txs in/out
func Init(port string) error {

	LogFile, _ = os.OpenFile("log/p2p "+time.Now().String(), os.O_RDWR|os.O_CREATE, 0666)
	log.SetOutput(LogFile)

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
		log.Print("Start mining as a non-bootstrap node.")
		err := bootstrap()
		if err != nil {
			return err
		}
	} else {
		log.Print("Start mining as a bootstrap node.")
	}

	go listener(port)
	return nil
}

func bootstrap() error {
	//connect to bootstrap server
	//initiate MINER_PING
	//add to connection list
	var conn net.Conn

	conn,err := net.Dial("tcp", "127.0.0.1:8000")

	if err != nil {
		fmt.Printf("%v\n", err)
		return err
	}

	packet := BuildPacket(MINER_PING,nil)
	conn.Write(packet)

	reader := bufio.NewReader(conn)
	header := ExtractHeader(reader)

	if header.TypeID == MINER_PONG {
		go minerConn(conn)
	} else {
		return errors.New("Connecting to bootstrap server failed.")
	}
	return nil
}

func listener(port string) {

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Printf("%v\n", err)
		return
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("%v\n", err)
			continue
		}
		go handleNewConn(conn)
	}
}

func handleNewConn(conn net.Conn) {

	log.Printf("New incoming connection: %v\n", conn.RemoteAddr().String())
	reader := bufio.NewReader(conn)
	header := ExtractHeader(reader)
	payload := make([]byte, header.Len-HEADER_LEN)
	var err error

	for cnt := 0; cnt < int(header.Len); cnt++ {
		payload[cnt], err = reader.ReadByte()
		if err != nil {
			log.Printf("%v\n", err)
			conn.Close()
			return
		}
	}

	processRequest(conn, header, payload)

	if header.TypeID == MINER_PING {
		go minerConn(conn)
	} else {
		conn.Close()
	}
}

func processRequest(conn net.Conn, header *Header, payload []byte) {

	log.Printf("%v: Received request with following header:\n%v", conn.RemoteAddr().String(), header)
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
	}
}

func receiveDataFromMiner() {

	for {
		select {
		case block := <-BlockOut:
			log.Printf("Received a block from the miner for broadcasting: %v\n", block)
			toBrdcst := BuildPacket(BLOCK_BRDCST, block)
			brdcstMsg <- toBrdcst
		case txInfo := <-TxsOut:
			log.Printf("Received a transaction from the miner for broadcasting: ID: %v, Payload: %v\n", txInfo.TxType,txInfo.Payload)
			toBrdcst := BuildPacket(txInfo.TxType, txInfo.Payload)
			brdcstMsg <- toBrdcst
		}
	}
}

//we can't broadcast incoming messages directly, need to forward them to the miner (to check if
//the tx has already been broadcast before, whether it was a valid tx at all)
func forwardTxToMiner(conn net.Conn, payload []byte, brdcstType uint8) {
	log.Printf("Received a transaction (ID: %v) from %v.\n", brdcstType,conn.RemoteAddr().String())
	TxsIn <- TxInfo{brdcstType, payload}
}
func forwardBlockToMiner(conn net.Conn, payload []byte) {
	log.Printf("Received a block from %v.\n", conn.RemoteAddr().String())
	BlockIn <- payload
}

//this is not accessed concurrently, one single goroutine
func broadcastService() {
	log.Print("Start broadcasting service.")
	for {
		select {
		//broadcasting all messages
		case msg := <-brdcstMsg:
			for p := range peers {
				p <- msg
			}
		case p := <-register:
			peers[p] = true
		case p := <-disconnect:
			delete(peers, p)
			close(p)
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

	ch := make(chan []byte)
	go peerWriter(conn, ch)

	log.Printf("Adding a new miner: %v\n", conn.RemoteAddr().String())

	register <- ch
	connReader := bufio.NewReader(conn)

	for {
		header := ExtractHeader(connReader)
		if header == nil {
			log.Printf("%v: Received corrupted header, closing connection.\n", conn.RemoteAddr().String())
			disconnect <- ch
			break
		}

		var err error
		payload := make([]byte, header.Len)
		for cnt := 0; cnt < int(header.Len); cnt++ {
			payload[cnt], err = connReader.ReadByte()
			if err != nil {
				log.Printf("%v: Peer disconnected (%v)\n", conn.RemoteAddr().String(), err)
				disconnect <- ch
				break
			}
		}

		processRequest(conn, header, payload)
	}
	conn.Close()
}

//will be our broadcast mechanism
func peerWriter(conn net.Conn, ch <-chan []byte) {
	for msg := range ch {
		log.Printf("Sending payload to %v\n", conn.RemoteAddr().String())
		conn.Write(msg)
	}
}
