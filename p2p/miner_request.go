package p2p

import (
	"errors"
)

//all the request in this file are specifically initiated by the miner package
func BlockReq(hash [32]byte) error {

	p := peers.getRandomPeer()
	if p == nil {
		return errors.New("Couldn't get a connection, request not transmitted.")
	}

	packet := BuildPacket(BLOCK_REQ, hash[:])
	sendData(p, packet)
	return nil
}

func TxReq(hash [32]byte, reqType uint8) error {

	p := peers.getRandomPeer()
	if p == nil {
		return errors.New("Couldn't get a connection, request not transmitted.")
	}

	packet := BuildPacket(reqType, hash[:])
	sendData(p, packet)

	return nil
}
