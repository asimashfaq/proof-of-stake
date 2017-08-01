package p2p

import (
	"time"
	"encoding/binary"
	"sync"
	"sort"
)

const (
	MIN_PEERS_FOR_TIME = 3
)

var (
	systemTime int64
	systemTimeLock sync.Mutex
)

func getTime() []byte {

	var buf [8]byte
	time := time.Now().Unix()
	binary.BigEndian.PutUint64(buf[:], uint64(time))
	return buf[:]
}

//Needs to be accessible by the miner package
func ReadSystemTime() int64 {
	return systemTime
}

func writeSystemTime() {
	peerTimes := peers.getPeerTimes()

	//add our own time as well
	peerTimes = append(peerTimes, time.Now().Unix())

	var ipeerTimes []int
	//remove all 0s and cast to int (needed to leverage sort.Ints)
	for _,time := range peerTimes {
		if time != 0 {
			ipeerTimes = append(ipeerTimes, int(time))
		}
	}

	//If we don't have at least MIN_PEERS_FOR_TIME different time values, we take our own system time for reference
	if len(ipeerTimes) < MIN_PEERS_FOR_TIME {
		systemTime = time.Now().Unix()
		return
	}

	systemTime = calcMedian(ipeerTimes)
}

func calcMedian(ipeerTimes []int) (median int64) {

	sort.Ints(ipeerTimes)
	//odd number of entries
	if len(ipeerTimes) % 2 == 1 {
		return int64(ipeerTimes[len(ipeerTimes)/2])
	} else {
		//even number of entries
		low := int64(ipeerTimes[len(ipeerTimes)/2])
		high := int64(ipeerTimes[len(ipeerTimes)/2+1])

		return (high+low)/2
	}
}