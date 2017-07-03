package p2p

import (
	"bufio"
)

func rcvData(p *peer) (*Header, []byte, error) {

	reader := bufio.NewReader(p.conn)
	header, err := ExtractHeader(reader)
	if err != nil {
		logger.Printf("Invalid packet received (%v)\n", err)
		p.conn.Close()
		return nil,nil,err
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