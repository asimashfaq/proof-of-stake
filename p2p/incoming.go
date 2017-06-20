package p2p

import (
	"bufio"
	"encoding/binary"
	"log"
	"net"
)

const (
	MAX_BUF_SIZE = 10000000 //10MB
	HEADER_LEN   = 7
	VERSION_ID   = 1
)

type Header struct {
	Len      uint32
	TypeID   uint8
	Version  uint8
	Reserved uint8
}


func handleConn(conn net.Conn) {

	reader := bufio.NewReader(conn)
	header := ExtractHeader(reader)

	if header.Len > MAX_BUF_SIZE {
		log.Printf("Input Size too large (%v), max. is %v\n", header.Len, MAX_BUF_SIZE)
	}
	inputBuf := make([]byte, header.Len)
	for i := 0; i < int(header.Len); i++ {

		in, err := reader.ReadByte()
		if err != nil {
			log.Printf("Error while reading the payload (%v)\n", err)
			break
		}
		inputBuf[i] = in
	}
	//not sure if reset has any benefits
	reader.Reset(conn)

	//certain inputs need to generate an appropriate output
	parseInput(conn, header, inputBuf)
}

func ConstructHeader(size int, typeID uint8) (header [HEADER_LEN]byte) {

	var len [4]byte
	binary.BigEndian.PutUint32(len[:], uint32(size))
	copy(header[0:4], len[:])
	header[4] = byte(typeID)
	header[5] = VERSION_ID
	return header
}

func ExtractHeader(reader *bufio.Reader) *Header {
	//the first four bytes of any incoming messages is the length of the payload
	//error catching after every read is necessary to avoid panicking
	var headerArr [HEADER_LEN]byte

	//reading byte by byte is surprisingly fast and works a lot better for concurrent connections
	for i := range headerArr {
		extr, err := reader.ReadByte()
		if err != nil {
			log.Printf("Invalid packet received (%v)\n", err)
			return nil
		}
		headerArr[i] = extr
	}

	lenBuf := [4]byte{headerArr[0], headerArr[1], headerArr[2], headerArr[3]}

	packetLen := binary.BigEndian.Uint32(lenBuf[:])

	header := new(Header)
	header.Len = packetLen
	header.TypeID = uint8(headerArr[4])
	header.Version = uint8(headerArr[5])
	header.Reserved = uint8(headerArr[6])
	return header
}

func parseInput(conn net.Conn, header *Header, data []byte) {

	//inspect header
	//parse input (what kind of tx, block etc.)
	switch header.TypeID {
	case ACCTX:
//		InAccTx(data)
	case FUNDSTX:
//		InFundsTx(data)
	case BLOCK:
		//InBlock(data)
	case TIME_REQ:
		timeRes(conn)
	case ACC_REQ:
		accRes(conn, data)
	}
	conn.Close()
}
