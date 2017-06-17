package main

import (
	"github.com/lisgie/bazo_miner/miner"
	"github.com/lisgie/bazo_miner/storage"
)

func main() {

	storage.Init()
	miner.Sync()
	go miner.Init()
	miner.InitSystem()
}
