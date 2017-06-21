package p2p

import (
	"net"
	"bufio"
	"log"
	"time"
	"fmt"
)

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

	/*ln, _ := net.Listen("tcp", ":"+PORT)

	for {
		conn, _ := ln.Accept()
		//creating new goroutine for every incoming request, not sure if smartest way to do it
		go handleConn(conn)
	}*/



	activePeers = make(map[string]*peer)
	//testing peer

}

func disconnectPeer(p *peer) {
	//clean up
	//I think garbage collector realises that it can remove p after removing it from the map?

	//delete(activePeers,p.conn.RemoteAddr().String())

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

func simpleFunc() {
	a,b := network.NeighborReq()
	fmt.Printf("%v, %v\n", a,b)
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
}