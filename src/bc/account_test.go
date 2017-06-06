package bc

import (
	"testing"
	"reflect"
)

func TestAccountSerialization(t *testing.T) {

	encodedAcc := EncodeAcc(accA)
	compareAcc := DecodeAcc(encodedAcc)
	if !reflect.DeepEqual(accA, compareAcc) {
		t.Error("Account Serialization failed!")
	}
}