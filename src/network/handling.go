package network

import (
	"bc"
	"net"
	"bufio"
)

func Init() {
	//for now mock data
	//will be later exchanged with listening on a socket

	ln, _ := net.Listen("tcp", ":8081")

	for {
		conn, _ := ln.Accept()
		//creating new goroutine for every incoming request, not sure if smartest way to do it
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {

	var input []byte
	input = make([]byte, 1000)
	reader := bufio.NewReader(conn)
	reader.Read(input)
	parseInput(input[1:])
	reader.Reset(conn)

	conn.Close()
}


func parseInput(data []byte) {

	//inspect header
	//parse input (what kind of tx, block etc.)
	switch data[0] {
	case ACCTX:
		bc.InAccTx(data[1:])
	case FUNDSTX:
		bc.InFundsTx(data[1:])
	case BLOCK:
	}
}

