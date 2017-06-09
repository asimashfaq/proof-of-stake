package main

import (
	"bc"
	"storage"
)

func main() {

	storage.Init()
	bc.Sync()
	go bc.Init()
	bc.InitSystem()

}