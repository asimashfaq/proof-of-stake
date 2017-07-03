package p2p

import (
	"math/rand"
	"time"
)

//get a random miner connection
func getRandomPeer() *peer {

	var peerSlice []*peer

	for {
		if len(peers) > 0 {
			break
		}
		time.Sleep(50*time.Millisecond)
	}

	pos := int(rand.Uint32()) % len(peers)
	for tmpPeer := range peers {
		peerSlice = append(peerSlice, tmpPeer)
	}

	return peerSlice[pos]
}

//needs to be accessible by the miner package, thus capital
func BlockReq(hash [32]byte) {

}

//asynchronous call, don't wait for response
func neighborReq() {

	p := getRandomPeer()
	if p == nil {
		return
	}

	packet := BuildPacket(NEIGHBOR_REQ, nil)
	sendData(p,packet)
}
