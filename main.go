package main

import (
	"os"
	"github.com/lisgie/bazo_miner/storage"
	"github.com/lisgie/bazo_miner/p2p"
	"github.com/lisgie/bazo_miner/miner"
	"fmt"
)

func main() {

	/*storage.Init()
	miner.Sync()
	go miner.Init()
	miner.InitSystem()*/

	var port,dbname string

	dbname = os.Args[1]
	port = os.Args[2]

	storage.Init(dbname)
	storage.DeleteAll()
	err := p2p.Init(port)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	miner.Init()
}
