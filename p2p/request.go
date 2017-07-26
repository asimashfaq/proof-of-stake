package p2p

func neighborReq() {

	p := getRandomPeer()
	if p == nil {
		logger.Print("Could not fetch a random peer.\n")
		return
	}

	packet := BuildPacket(NEIGHBOR_REQ, nil)
	sendData(p, packet)
}
