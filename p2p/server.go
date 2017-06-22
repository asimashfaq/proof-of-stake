package p2p

import (
	"net"
)

var (
	peers map[peer]bool
	brdcstMsg chan []byte
	register chan peer
	disconnect chan peer
)

type peer chan<- []byte

func Init() {

	//after this call, there are some peers connected

	//to avoid that all new peers connect to all bootstrap peers, we just connect to one or two
	//and request neighboring ip addresses
	//neighboring ip addresses are incoming and outgoing addresses

	go handleEvents()
	go checkHealth()
	listenIncoming()
}

func isMiner(conn net.Conn) bool {

	//send ping request, abort conn if no pong comes back
	conn.Close()
	return false
}

func listenIncoming() {

	//listen and spawn handleConn

}


func Broadcast(payload []byte) {

	//did we already broadcast it before?

	brdcstMsg<-payload
}

//this is not accessed concurrently, one single goroutine
func handleEvents() {

	for {
		select {
		//broadcasting all messages
		case msg := <-brdcstMsg:
			for p := range peers {
				p<-msg
			}
		case p := <-register:
			peers[p] = true
		case p := <-disconnect:
			delete(peers,p)
			close(p)
		}
	}
}

//single goroutine that makes sure that system is well connected
func checkHealth() {

	for {
		break
		//initiate new connection if not enough
		//and call go outgoingConn(conn)
	}
}

func outgoingConn(conn net.Conn) {

	ch := make(chan []byte)
	go clientWriter(conn,ch)

	register<-ch

	var buf []byte
	for {
		_,err := conn.Read(buf)
		if err != nil {
			//remote end has disconnected
			disconnect<-ch
			break
		}
	}
	conn.Close()
}

func clientWriter(conn net.Conn, ch <-chan []byte) {
	for msg := range ch {
		conn.Write(msg)
	}
}


/*
const(
	//each peer is connected to this many peers to send broadcasts and unicasts to
	OUT_CONN_LIMIT = 8
	PORT = "8000"
)

type network_iface interface {
	NeighborReq() ([]string, error)
}

var (
	activePeers map[string]*peer
	potentialPeers []string
	network network_iface
)

type peer struct {
	conn net.Conn
	readwriter bufio.ReadWriter
}

func setDebug(iface network_iface) {
	network = iface
}

func Init() {

	network = production{}



	activePeers = make(map[string]*peer)
	//testing peer

}

func disconnectPeer(p *peer) {
	//clean up
	//I think garbage collector realises that it can remove p after removing it from the map?

	delete(activePeers,p.conn.RemoteAddr().String())

	for ;; {

		if len(activePeers) >= OUT_CONN_LIMIT {
			break
		}

		addr := getNewAddress()

		//initiate connection with received message
		newConn,err := net.Dial("tcp",addr)
		if err != nil {
			log.Printf("Couldn't initiate connection to IP Address: %v\n", addr)
			continue
		} else {
			var newReaderWriter bufio.ReadWriter
			newReaderWriter.Reader.Reset(newConn)
			newReaderWriter.Writer.Reset(newConn)
			p := &peer{
				conn: newConn,
				readwriter: newReaderWriter,
			}
			activePeers[newConn.RemoteAddr().String()] = p
		}
	}
}

func getNewAddress() (string) {

	var addrList []string
	var err error

	for {
		if len(potentialPeers) == 0 {
			addrList,err = network.NeighborReq()
			if err != nil {
				log.Printf("%v\n", err)
				continue
			}
			//remove duplicates and already active peers
			checkDuplicates(addrList)
		}

		if len(potentialPeers) > 0 {
			break
		}
		time.Sleep(200*time.Millisecond)
	}

	addr := potentialPeers[len(potentialPeers)-1]
	potentialPeers = potentialPeers[:len(potentialPeers)-1]
	return addr
}

func checkDuplicates(addrList []string) {

	for _,newAddr := range addrList {
		if _,exists := activePeers[newAddr]; exists {
			continue
		}

		var duplicate bool
		for _,existingAddr := range potentialPeers {
			if newAddr == existingAddr {
				duplicate = true
				break
			}
		}
		if !duplicate {
			potentialPeers = append(potentialPeers,newAddr)
		}
	}
}*/