package p2p

import (
	"bufio"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

const (
	PORT       = 8000
	MIN_MINERS = 10
	MAX_MINERS = 20
	TX_BUFFER = 10
)

var (
	LogFile    *os.File
	peers      map[peer]bool
	brdcstMsg  chan []byte
	register   chan peer
	disconnect chan peer

	TxsIn chan TxInfo
	BlockIn chan []byte

	TxsOut chan TxInfo
	BlockOut chan []byte
)

//we need to decode incoming transactions, therefore type is needed
//for outgoing transactions, the p2p package needs the information to build the proper header
type TxInfo struct {
	TxType uint8
	Payload []byte
}

type peer chan<- []byte

//4 channels for communication with miner, blocks in/out and txs in/out
func Init() {

	TxsIn = make(chan TxInfo, TX_BUFFER)
	BlockIn = make(chan []byte)
	TxsOut = make(chan TxInfo, TX_BUFFER)
	BlockOut = make(chan []byte)

	peers = make(map[peer]bool)
	brdcstMsg = make(chan []byte)
	register = make(chan peer)
	disconnect = make(chan peer)

	go broadcastService()
	//go checkHealth()

	LogFile, _ = os.OpenFile("logp2p "+time.Now().String(), os.O_RDWR|os.O_CREATE, 0666)
	log.SetOutput(LogFile)
	//after this call, there are some peers connected

	//to avoid that all new peers connect to all bootstrap peers, we just connect to one or two
	//and request neighboring ip addresses
	//neighboring ip addresses are incoming and outgoing addresses

	listener, err := net.Listen("tcp", ":"+strconv.Itoa(PORT))
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

	log.Printf("%v: Received request with following header:\n%v", conn.RemoteAddr().String(),header)
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

func initiateBlockBroadcast(conn net.Conn, payload []byte) {


}

func receiveDataFromMiner() {

	for {
		select {
		case block := <-BlockOut:
			toBrdcst := BuildPacket(BLOCK_BRDCST, block)
			brdcstMsg <- toBrdcst
		case txInfo := <-TxsOut:
			toBrdcst := BuildPacket(txInfo.TxType,txInfo.Payload)
			brdcstMsg<-toBrdcst
		}
	}
}

//we can't broadcast incoming messages directly, need to forward them to the miner (to check if
//the tx has already been broadcast before, whether it was a valid tx at all)
func forwardTxToMiner(conn net.Conn, payload []byte, brdcstType uint8) { TxsIn<-TxInfo{brdcstType, payload} }
func forwardBlockToMiner(conn net.Conn, payload []byte) { BlockIn<-payload }

//this is not accessed concurrently, one single goroutine
func broadcastService() {
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
			time.Sleep(10*time.Second)
			continue
		}
		//initiate new connection if not enough
		//and call go outgoingConn(conn)
	}
}

func minerConn(conn net.Conn) {

	ch := make(chan []byte)
	go peerWriter(conn, ch)

	log.Printf("%v: Adding a new miner\n", conn.RemoteAddr().String())

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
		conn.Write(msg)
	}
}
