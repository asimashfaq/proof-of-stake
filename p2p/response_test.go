package p2p

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"testing"
)

func Test_NeighborRes(t *testing.T) {

	ipportList := []string{
		"127.0.0.1:8000",
		"127.0.0.1:8005",
		"127.0.0.1:40000",
	}

	payload := _neighborRes(ipportList)
	fmt.Printf("%v\n", payload)

	//Check for correct deserialization
	index := 0
	if payload[index] != 127 || payload[index+1] != 0 || payload[index+2] != 0 || payload[index+3] != 1 ||
		strconv.Itoa(int(binary.BigEndian.Uint16(payload[index+4:index+6]))) != "8000" {
		t.Error("IP/Port Deserialization failed.")
	}

	index += 6
	if payload[index] != 127 || payload[index+1] != 0 || payload[index+2] != 0 || payload[index+3] != 1 ||
		strconv.Itoa(int(binary.BigEndian.Uint16(payload[index+4:index+6]))) != "8005" {
		t.Error("IP/Port Deserialization failed.")
	}

	index += 6
	if payload[index] != 127 || payload[index+1] != 0 || payload[index+2] != 0 || payload[index+3] != 1 ||
		strconv.Itoa(int(binary.BigEndian.Uint16(payload[index+4:index+6]))) != "40000" {
		t.Error("IP/Port Deserialization failed.")
	}
}
