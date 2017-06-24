package main

import (
	"github.com/lisgie/bazo_miner/p2p"
	"github.com/lisgie/bazo_miner/storage"
)

func main() {

	/*storage.Init()
	miner.Sync()
	go miner.Init()
	miner.InitSystem()*/

	storage.Init()
	p2p.Init()
}
