package p2p

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"log"
)

const (
	MAX_BUF_SIZE = 10000000 //10MB
	HEADER_LEN   = 5
	VERSION_ID   = 1
)

//Java can't handle uints, should we only allow lengths of up to 2^31?
type Header struct {
	Len    uint32
	TypeID uint8
}

func BuildPacket(typeID uint8, payload []byte) (packet []byte) {

	var payloadLen [4]byte
	packet = make([]byte, HEADER_LEN+len(payload))
	binary.BigEndian.PutUint32(payloadLen[:], uint32(len(payload)))
	copy(packet[0:4], payloadLen[:])
	packet[4] = byte(typeID)
	copy(packet[5:], payload)
	return packet
}

func ExtractHeader(reader *bufio.Reader) (*Header,error) {
	//the first four bytes of any incoming messages is the length of the payload
	//error catching after every read is necessary to avoid panicking
	var headerArr [HEADER_LEN]byte
	//reading byte by byte is surprisingly fast and works a lot better for concurrent connections
	for i := range headerArr {
		extr, err := reader.ReadByte()
		if err != nil {
			log.Printf("Invalid packet received (%v)\n", err)
			return nil,err
		}
		headerArr[i] = extr
	}

	lenBuf := [4]byte{headerArr[0], headerArr[1], headerArr[2], headerArr[3]}

	packetLen := binary.BigEndian.Uint32(lenBuf[:])

	header := new(Header)
	header.Len = packetLen
	header.TypeID = uint8(headerArr[4])
	return header,nil
}

func (header Header) String() string {
	return fmt.Sprintf(
		"Length: %v\n"+
			"TypeID: %v\n",
		header.Len,
		header.TypeID,
	)
}
