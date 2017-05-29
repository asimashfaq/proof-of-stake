package bc

import (
	"testing"
)

func TestAccTx(t *testing.T) {

	//creating some root-signed new accounts
	for i := 0; i <= 1000; i++ {
		tx, _ := ConstrAccTx(&RootPrivKey)
		if tx.verify() == false {
			t.Errorf("AccTx could not be verified: %v\n", tx)
		}
	}
}