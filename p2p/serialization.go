package p2p

import (
	"bufio"
	"net"
)

func rcvData(conn net.Conn) (*Header, []byte, error) {

	reader := bufio.NewReader(conn)
	header := ExtractHeader(reader)
	payload := make([]byte, header.Len-HEADER_LEN)

	var err error
	for cnt := 0; cnt < int(header.Len); cnt++ {
		payload[cnt], err = reader.ReadByte()
		if err != nil {
			return nil, nil, err
		}
	}

	return header, payload, nil
}
