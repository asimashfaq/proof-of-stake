package main

import (
	"github.com/lisgie/bazo_miner/bc"
	"github.com/lisgie/bazo_miner/storage"
)

func main() {

	storage.Init()
	bc.Sync()
	go bc.Init()
	bc.InitSystem()
}
