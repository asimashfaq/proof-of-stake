package main

import (
	"bc"
	"net"
	"bufio"
)

func listenForData() {
	//for now mock data
	//will be later exchanged with listening on a socket

	ln, _ := net.Listen("tcp", ":8081")
	conn, _ := ln.Accept()
	var input []byte
	input = make([]byte, 100)

	for {
		reader := bufio.NewReader(conn)
		reader.Read(input)
		bc.ProcessInput(input)
		reader.Reset(conn)
	}
	conn.Close()
}

func setUpTestSystem() {


}

func main() {

	bc.Sync()
	go listenForData()
	bc.InitSystem()
	//setUpTestSystem()

	/*_rootPub1,_ := new(big.Int).SetString(bc.RootPub1,16)
	_rootPub2,_ := new(big.Int).SetString(bc.RootPub2,16)
	_rootPriv,_ := new(big.Int).SetString(bc.RootPriv,16)
	rootPubKey := ecdsa.PublicKey{
		elliptic.P256(),
		_rootPub1,
		_rootPub2,
	}
	rootPrivKey := ecdsa.PrivateKey{
		rootPubKey,
		_rootPriv,
	}
	var rootPubKeyHash [32]byte
	var tmp [64]byte
	copy(tmp[:32],_rootPub1.Bytes())
	copy(tmp[32:],_rootPub2.Bytes())
	rootPubKeyHash = sha3.Sum256(tmp[:])*/
}