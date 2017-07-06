package p2p

import (
	"bufio"
	"math/rand"
	"time"
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

	var peerSlice []*peer

	for {
		if len(peers) > 0 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	pos := int(rand.Uint32()) % len(peers)
	for tmpPeer := range peers {
		peerSlice = append(peerSlice, tmpPeer)
	}

	return peerSlice[pos]
}
