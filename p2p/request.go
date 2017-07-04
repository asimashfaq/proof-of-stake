package p2p

//asynchronous call, don't wait for response
func neighborReq() {

	p := getRandomPeer()
	if p == nil {
		return
	}

	packet := BuildPacket(NEIGHBOR_REQ, nil)
	sendData(p,packet)
}

