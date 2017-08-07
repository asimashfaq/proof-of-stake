package main

import (
	"fmt"
	"github.com/lisgie/bazo_miner/miner"
	"github.com/lisgie/bazo_miner/p2p"
	"github.com/lisgie/bazo_miner/storage"
	"os"
)

func main() {

	var localConn, dbname string

	dbname = os.Args[1]
	localConn = os.Args[2]

	storage.Init(dbname)
	storage.DeleteAll()
	err := p2p.Init(localConn)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	miner.Init()
}
