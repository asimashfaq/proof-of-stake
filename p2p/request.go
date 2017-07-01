package p2p

import (
	"math/rand"
	"net"
)

//get a random miner connection
func getConn() net.Conn {

	var conns []net.Conn

	if len(peers) == 0 {
		return nil
	}

	pos := int(rand.Uint32()) % len(peers)
	for tmpConn := range peers {
		conns = append(conns, tmpConn.conn)
	}

	return conns[pos]
}

func neighborReq() []byte {

	conn := getConn()
	if conn == nil {
		return nil
	}

	packet := BuildPacket(NEIGHBOR_REQ, nil)
	conn.Write(packet)

	header, payload, err := rcvData(conn)
	if err != nil || header.TypeID != NEIGHBOR_RES {
		return nil
	}

	return payload
}
