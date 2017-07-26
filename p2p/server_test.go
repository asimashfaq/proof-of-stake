package p2p

import (
	"testing"
)

func TestCheckDuplicates(t *testing.T) {

	/*	activePeers["1.1.1.1:8000"] = new(peer)
		activePeers["1.1.1.2:8000"] = new(peer)
		activePeers["1.1.1.3:8000"] = new(peer)
		activePeers["1.1.1.4:8000"] = new(peer)

		potentialPeers = append(potentialPeers,"2.2.2.1:8000")
		potentialPeers = append(potentialPeers,"2.2.2.2:8000")
		potentialPeers = append(potentialPeers,"2.2.2.3:8000")
		potentialPeers = append(potentialPeers,"2.2.2.4:8000")

		var addrList []string
		addrList = append(addrList,"1.1.1.1:8000")
		addrList = append(addrList,"4.4.4.4:8000") //will be added
		addrList = append(addrList,"2.2.2.2:8000")
		addrList = append(addrList,"3.3.3.3:8000") //will be added

		tmpNumberPeers := len(potentialPeers)
		checkDuplicates(addrList)
		if tmpNumberPeers+2 != len(potentialPeers) {
			t.Error("Checking duplicates was not successful.")
		}*/
}

func TestGetNewAddress(t *testing.T) {

	//dependent on neighorReq, will be implemented later
	/*Init()


	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockObj := mock_p2p.NewMocknetwork_iface(mockCtrl)
	mockObj.EXPECT().NeighborReq().Return([]string{"123"},nil)
	setDebug(mockObj)

	simpleFunc()*/

	//fmt.Printf("%v\n", getNewAddress())

}
