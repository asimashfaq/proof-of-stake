package p2p

import (
	"bufio"
	"math/rand"
	"encoding/binary"
)

func rcvData(p *peer) (*Header, []byte, error) {

	reader := bufio.NewReader(p.conn)
	header, err := ExtractHeader(reader)
	if err != nil {
		logger.Printf("Connection to %v aborted: (%v)\n", p.conn.RemoteAddr().String(), err)
		p.conn.Close()
		return nil, nil, err
	}
	payload := make([]byte, header.Len)

	for cnt := 0; cnt < int(header.Len); cnt++ {
		payload[cnt], err = reader.ReadByte()
		if err != nil {
			p.conn.Close()
			return nil, nil, err
		}
	}

	return header, payload, nil
}

func sendData(p *peer, payload []byte) {
	p.l.Lock()
	p.conn.Write(payload)
	p.l.Unlock()
}

//get a random miner connection
func getRandomPeer() *peer {

	peers.peerMutex.Lock()
	defer peers.peerMutex.Unlock()
	if len(peers.peerConns) == 0 {
		return nil
	}

	var peerSlice []*peer

	pos := int(rand.Uint32()) % len(peers.peerConns)
	for tmpPeer := range peers.peerConns {
		peerSlice = append(peerSlice, tmpPeer)
	}

	return peerSlice[pos]
}

//We have to prevent to connect to miners twice
func peerExists(newIpport string) bool {

	allConns := peers.getAll()

	for _,ipport := range allConns {
		if ipport == newIpport {
			return true
		}
	}

	return false
}

func peerSelfConn(newIpport string) bool {
	return newIpport == localConn
}

func BuildPacket(typeID uint8, payload []byte) (packet []byte) {

	logger.Printf("Building new packet with type ID (%v) and packet length (%v).\n", typeID, len(payload))
	var payloadLen [4]byte
	packet = make([]byte, HEADER_LEN+len(payload))
	binary.BigEndian.PutUint32(payloadLen[:], uint32(len(payload)))
	copy(packet[0:4], payloadLen[:])
	packet[4] = byte(typeID)
	copy(packet[5:], payload)
	return packet
}

func ExtractHeader(reader *bufio.Reader) (*Header, error) {
	//the first four bytes of any incoming messages is the length of the payload
	//error catching after every read is necessary to avoid panicking
	var headerArr [HEADER_LEN]byte
	//reading byte by byte is surprisingly fast and works a lot better for concurrent connections
	for i := range headerArr {
		extr, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}
		headerArr[i] = extr
	}

	lenBuf := [4]byte{headerArr[0], headerArr[1], headerArr[2], headerArr[3]}

	packetLen := binary.BigEndian.Uint32(lenBuf[:])

	header := new(Header)
	header.Len = packetLen
	header.TypeID = uint8(headerArr[4])
	return header, nil
}