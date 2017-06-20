package p2p

import (
	"encoding/binary"
	"net"
	"time"
	"github.com/lisgie/bazo_miner/protocol"
)

func prepareRequest(typeID uint8) (responseData []byte) {

	//conn := getPeer()

	return responseData
}

func blockReq(hash [32]byte) (b *protocol.Block) {

	return b
}

func timeRes(conn net.Conn) {

	var buf [8]byte
	time := time.Now().Unix()
	binary.BigEndian.PutUint64(buf[:], uint64(time))
	toSend := make([]byte, len(buf)+HEADER_LEN)
	header := ConstructHeader(len(buf), TIME_RES)
	copy(toSend[0:HEADER_LEN], header[:])
	copy(toSend[HEADER_LEN:], buf[:])
	conn.Write(toSend)
}

func accRes(conn net.Conn, data []byte) {

	/*var hash [32]byte
	copy(hash[:], data[0:32])
	acc := getAccountFromHash(hash)
	encodedAcc := EncodeAcc(acc)
	header := ConstructHeader(len(encodedAcc), ACC_RES)
	toSend := make([]byte, len(header)+len(encodedAcc))
	copy(toSend[:HEADER_LEN], header[:])
	copy(toSend[HEADER_LEN:], encodedAcc)
	conn.Write(toSend)*/
}

func neighborReq() ([]string,error) {

	return []string{
		"127.0.0.1:8005",
		"127.0.0.1:8006",
		"127.0.0.1:8007",
		"127.0.0.1:8008",
	},nil
}