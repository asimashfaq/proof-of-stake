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
		header := ConstructHeader(0, NOT_FOUND)
		toSend := make([]byte, HEADER_LEN)
		copy(toSend[:], header[:])
		conn.Write(toSend)
		return
	}

	header := ConstructHeader(len(tx), txKind)
	toSend := make([]byte, HEADER_LEN+len(tx))
	copy(toSend[:HEADER_LEN], header[:])
	copy(toSend[HEADER_LEN:], tx)
	conn.Write(toSend)
}

func blockRes(conn net.Conn, payload []byte) {

	var blockHash [32]byte
	copy(blockHash[:], payload[0:32])

	block := storage.ReadBlock(blockHash)

	if block == nil {
		header := ConstructHeader(0, NOT_FOUND)
		toSend := make([]byte, HEADER_LEN)
		copy(toSend[:], header[:])
		conn.Write(toSend)
		return
	}

	header := ConstructHeader(len(block), BLOCK_RES)
	toSend := make([]byte, HEADER_LEN+len(block))
	copy(toSend[:HEADER_LEN], header[:])
	copy(toSend[HEADER_LEN:], block)
	conn.Write(toSend)
}

func accRes(conn net.Conn, payload []byte) {

	var hash [32]byte
	copy(hash[:], payload[0:32])
	acc := storage.GetAccountFromHash(hash)
	encodedAcc := acc.Encode()

	if encodedAcc == nil {
		header := ConstructHeader(0, NOT_FOUND)
		toSend := make([]byte, HEADER_LEN)
		copy(toSend[:], header[:])
		conn.Write(toSend)
		return
	}

	header := ConstructHeader(len(encodedAcc), ACC_RES)
	toSend := make([]byte, len(header)+len(encodedAcc))
	copy(toSend[:HEADER_LEN], header[:])
	copy(toSend[HEADER_LEN:], encodedAcc)
	conn.Write(toSend)
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

func pongRes(conn net.Conn, payload []byte) {

	header := ConstructHeader(0, MINER_PONG)
	toSend := make([]byte, len(header))
	copy(toSend[:HEADER_LEN], header[:])
	conn.Write(toSend)
}
