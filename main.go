package main

import (
	"github.com/lisgie/bazo_miner/storage"
	"github.com/lisgie/bazo_miner/bc"
)

func main() {

	storage.Init()
	bc.Sync()
	go bc.Init()
	bc.InitSystem()
}
