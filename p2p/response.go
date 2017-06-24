package p2p

import (
	"encoding/binary"
	"github.com/lisgie/bazo_miner/storage"
	"net"
	"time"
)

func txRes(conn net.Conn, payload []byte, txKind uint8) {

	var txHash [32]byte
	copy(txHash[:], payload[0:32])

	var tx []byte
	//we need to look in the mempool as well as the validated txs
	openTx := storage.ReadOpenTx(txHash)
	closedTx := storage.ReadClosedTx(txHash)

	if openTx != nil {
		tx = openTx
	} else if closedTx != nil {
		tx = closedTx
	}

	if tx == nil {
		packet := BuildPacket(NOT_FOUND,nil)
		conn.Write(packet)
		return
	}

	packet := BuildPacket(txKind, payload)
	conn.Write(packet)
}

func blockRes(conn net.Conn, payload []byte) {

	var blockHash [32]byte
	copy(blockHash[:], payload[0:32])

	block := storage.ReadBlock(blockHash)

	if block == nil {
		packet := BuildPacket(NOT_FOUND, nil)
		conn.Write(packet)
		return
	}

	packet := BuildPacket(BLOCK_RES, block)
	conn.Write(packet)
}

func accRes(conn net.Conn, payload []byte) {

	var hash [32]byte
	copy(hash[:], payload[0:32])
	acc := storage.GetAccountFromHash(hash)
	encodedAcc := acc.Encode()

	if encodedAcc == nil {
		packet := BuildPacket(NOT_FOUND,nil)
		conn.Write(packet)
		return
	}
	packet := BuildPacket(ACC_RES, encodedAcc)
	conn.Write(packet)
}

func timeRes(conn net.Conn) {

	var buf [8]byte
	time := time.Now().Unix()
	binary.BigEndian.PutUint64(buf[:], uint64(time))
	packet := BuildPacket(TIME_RES, buf[:])
	conn.Write(packet)
}

func pongRes(conn net.Conn, payload []byte) {

	packet := BuildPacket(MINER_PONG, nil)
	conn.Write(packet)
}
