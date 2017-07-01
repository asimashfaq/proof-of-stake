package p2p

import (
	"encoding/binary"
	"github.com/lisgie/bazo_miner/protocol"
	"github.com/lisgie/bazo_miner/storage"
	"net"
	"strconv"
	"strings"
	"time"
)

func txRes(conn net.Conn, payload []byte, txKind uint8) {

	var txHash [32]byte
	copy(txHash[:], payload[0:32])

	var tx protocol.Transaction
	//we need to look in the mempool as well as the validated txs
	openTx := storage.ReadOpenTx(txHash)
	closedTx := storage.ReadClosedTx(txHash)

	if openTx != nil {
		tx = openTx
	} else if closedTx != nil {
		tx = closedTx
	}

	if tx == nil {
		packet := BuildPacket(NOT_FOUND, nil)
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

	packet := BuildPacket(BLOCK_RES, block.Encode())
	conn.Write(packet)
}

func accRes(conn net.Conn, payload []byte) {

	var hash [32]byte
	copy(hash[:], payload[0:32])
	acc := storage.GetAccountFromHash(hash)
	encodedAcc := acc.Encode()

	if encodedAcc == nil {
		packet := BuildPacket(NOT_FOUND, nil)
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

func neighborRes(conn net.Conn, payload []byte) {
	//only supporting ipv4 addresses for now, makes fixed-size structure easier
	//in the future following structure is possible:
	//1) nr of ipv4 addresses, 2) nr of ipv6 addresses, followed by list of both

	var packet []byte

	payload = make([]byte, len(peers)*4) //4 = size of ipv4 address
	index := 0
	for p := range peers {
		discardPort := strings.Split(p.conn.RemoteAddr().String(), ":")
		split := strings.Split(discardPort[0], ".")
		//serializing ip addresses
		for ipv4addr := 0; ipv4addr < 4; ipv4addr++ {
			addrPart, err := strconv.Atoi(split[ipv4addr])
			if err != nil {
				packet = BuildPacket(NEIGHBOR_RES, nil)
				conn.Write(packet)
				return
			}
			payload[index] = byte(addrPart)
			index++
		}
	}

	packet = BuildPacket(NEIGHBOR_RES, payload)
	conn.Write(packet)
}
