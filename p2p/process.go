package p2p

import (
	"encoding/binary"
	"strconv"
)

func processTimeRes(p *peer, payload []byte) {

	time := int64(binary.BigEndian.Uint64(payload))
	//concurrent writes need to be protected
	//we use the same lock to prevent concurrent writes. It would be more efficient to use different locks
	//but the speedup is so marginal that it's not worth it
	p.l.Lock()
	defer p.l.Unlock()
	p.time = time
}

func processNeighborRes(p *peer, payload []byte) {

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
