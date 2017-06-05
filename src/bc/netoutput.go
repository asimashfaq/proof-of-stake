package bc

import (
	"net"
	"time"
	"encoding/binary"
)

func timeReq(conn net.Conn) {

	var buf [8]byte
	time := time.Now().Unix()
	binary.BigEndian.PutUint64(buf[:], uint64(time))
	toSend := make([]byte, len(buf)+HEADER_LEN)
	header := ConstructHeader(len(buf),TIME_RES)
	copy(toSend[0:HEADER_LEN],header[:])
	copy(toSend[HEADER_LEN:],buf[:])
	conn.Write(toSend)
}
