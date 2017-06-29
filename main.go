package main

import (
	"github.com/lisgie/bazo_miner/p2p"
	"github.com/lisgie/bazo_miner/storage"
	"github.com/lisgie/bazo_miner/miner"
)

func main() {

	/*storage.Init()
	miner.Sync()
	go miner.Init()
	miner.InitSystem()*/

	storage.Init()
	storage.DeleteAll()
	p2p.Init()
	miner.Init()
}
