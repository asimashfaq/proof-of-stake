package p2p

import (
	"encoding/binary"
	"strconv"
)

func processNeighborRes(p *peer, payload []byte) {

	logger.Printf("Received (%v) address(es) from peer with IP address: %v\n", len(payload)/(IPV4ADDR+PORT_SIZE), p.conn.RemoteAddr())

	//parse the incoming ipv4 addresses
	ipportList := _processNeighborRes(payload)

	for _, ipportIter := range ipportList {
		logger.Printf("IP/Port received: %v\n", ipportIter)
		iplistChan <- ipportIter
	}
}

//Decoupled for cleaner testing
func _processNeighborRes(payload []byte) (ipportList []string) {

	index := 0
	for cnt := 0; cnt < len(payload)/(IPV4ADDR+PORT_SIZE); cnt++ {
		var addr string
		for singleAddr := index; singleAddr < index+IPV4ADDR; singleAddr++ {
			tmp := int(payload[singleAddr])
			addr += strconv.Itoa(tmp)
			addr += "."
		}
		//remove trailing dot
		addr = addr[:len(addr)-1]
		addr += ":"
		//extract port number
		addr += strconv.Itoa(int(binary.BigEndian.Uint16(payload[index+4 : index+6])))

		//add ipaddr to the channel
		ipportList = append(ipportList, addr)
		index += IPV4ADDR + PORT_SIZE
	}
	return ipportList
}
