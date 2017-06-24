package p2p

import (
	"bufio"
	"github.com/lisgie/bazo_miner/protocol"
	"github.com/lisgie/bazo_miner/storage"
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
)

var (
	LogFile    *os.File
	peers      map[peer]bool
	brdcstMsg  chan msgPeer
	register   chan peer
	disconnect chan peer
)

//we need that in order to not send back the broadcast to the peer we received it from
type msgPeer struct {
	payload []byte
	conn    net.Conn
}

type peer chan<- msgPeer

func Init() {

	peers = make(map[peer]bool)
	brdcstMsg = make(chan msgPeer)
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
		initiateTxBroadcast(conn, payload, FUNDSTX_BRDCST)
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

//miner created a new block that needs to be broadcast
func MinerBroadcastBlock(payload []byte) { brdcstMsg <- msgPeer{payload, nil} }

func initiateBlockBroadcast(conn net.Conn, payload []byte) {

	var block *protocol.Block
	block = block.Decode(payload)
	if block == nil {
		return
	}
	if dec := storage.ReadBlock(block.Hash); dec != nil {
		return
	}
	toBrdcst := BuildPacket(BLOCK_BRDCST, payload)
	brdcstMsg <- msgPeer{toBrdcst, conn}
}

func initiateTxBroadcast(conn net.Conn, payload []byte, brdcstType uint8) {

	//check if we already did a broadcast by checking whether it's already been written to memory
	//we save memory by not transmitting the hash, this has to be calculated, worth the tradeoff I guess
	switch brdcstType {
	case FUNDSTX_BRDCST:
		var fTx *protocol.FundsTx
		fTx = fTx.Decode(payload)

		if fTx == nil {
			return
		}
		if dec := storage.ReadOpenTx(fTx.Hash()); dec != nil {
			return
		}
		if dec := storage.ReadClosedTx(fTx.Hash()); dec != nil {
			return
		}
	case ACCTX_BRDCST:
		var aTx *protocol.AccTx
		aTx = aTx.Decode(payload)
		if aTx == nil {
			return
		}
		if dec := storage.ReadOpenTx(aTx.Hash()); dec != nil {
			return
		}
		if dec := storage.ReadClosedTx(aTx.Hash()); dec != nil {
			return
		}
	case CONFIGTX_BRDCST:
		var cTx *protocol.ConfigTx
		cTx = cTx.Decode(payload)
		if cTx == nil {
			return
		}
		if dec := storage.ReadOpenTx(cTx.Hash()); dec != nil {
			return
		}
		if dec := storage.ReadClosedTx(cTx.Hash()); dec != nil {
			return
		}
	}

	//build new broadcast packet
	toBrdcst := BuildPacket(brdcstType,payload)
	brdcstMsg<-msgPeer{toBrdcst,conn}
}

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

	ch := make(chan msgPeer)
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
func peerWriter(conn net.Conn, ch <-chan msgPeer) {
	for msg := range ch {
		if conn != msg.conn {
			conn.Write(msg.payload)
		}
	}
}
