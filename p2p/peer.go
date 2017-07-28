package p2p

import (
	"sync"
	"net"
)

//The reason we use an additional listener port is because the port the miner connected to this peer
//is not the same as the one it listens to for new connections. When we are queried for neighbors
//we send the IP address in p.conn.RemotAddr() with the listenerPort
type peer struct {
	conn net.Conn
	ch   chan []byte
	l    sync.Mutex
	listenerPort string
}

type peersStruct struct {
	peerConns      map[*peer]bool
	peerMutex 	sync.Mutex
}

func (peers peersStruct) add(p *peer) {
	peers.peerMutex.Lock()
	defer peers.peerMutex.Unlock()
	peers.peerConns[p] = true
}

func (peers peersStruct) delete(p *peer) {
	peers.peerMutex.Lock()
	defer peers.peerMutex.Unlock()
	delete(peers.peerConns, p)
}

func (peers peersStruct) len() int {
	//being extra paranoid, probably not strictly needed because reading is thread-safe
	peers.peerMutex.Lock()
	defer peers.peerMutex.Unlock()
	return len(peers.peerConns)
}

func (peers peersStruct) getAll() []string {
	peers.peerMutex.Lock()
	defer peers.peerMutex.Unlock()
	return nil
}




