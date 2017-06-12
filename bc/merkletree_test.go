package bc

import (
	"testing"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
)

func TestBuildMerkleTree(t *testing.T) {

	var hashSlice [][32]byte
	var hash1,hash2,hash3 [32]byte
	var tmpHash []byte
	var tx, tx2, tx3 *fundsTx

	//generating a private key and prepare data
	privA,_ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	tx,_ = ConstrFundsTx(0x01,23,1,0,[32]byte{'0'},[32]byte{'1'},privA)
	tx2,_ = ConstrFundsTx(0x01,26,1,1,[32]byte{'0'},[32]byte{'1'},privA)
	tx3,_ = ConstrFundsTx(0x02,25,1,34,[32]byte{'2'},[32]byte{'5'}, privA)

	//testing with 1,2,3 nodes
	hash1 = hashFundsTx(tx)
	hash2 = hashFundsTx(tx2)
	hash3 = hashFundsTx(tx3)

	//test with one node
	//self hash
	tmpHash = append(hash1[:],hash1[:]...)
	hashSlice = append(hashSlice,hash1)
	if serializeHashContent(tmpHash) != buildMerkleTree(hashSlice) {
		t.Errorf("Hashes don't match: %x != %x\n", serializeHashContent(tmpHash), buildMerkleTree(hashSlice))
	}

	//two nodes
	tmpHash = append(hash1[:],hash2[:]...)
	hashSlice = append(hashSlice, hash2)
	if serializeHashContent(tmpHash) != buildMerkleTree(hashSlice) {
		t.Errorf("Hashes don't match: %x != %x\n", serializeHashContent(tmpHash), buildMerkleTree(hashSlice))
	}

	//three nodes
	tmpHash = append(hash1[:],hash2[:]...)
	tmpHashHash := serializeHashContent(tmpHash)
	tmpHash2 := append(hash3[:],hash3[:]...)
	tmpHashHash2 := serializeHashContent(tmpHash2)
	finalHash := append(tmpHashHash[:],tmpHashHash2[:]...)
	hashSlice = append(hashSlice,hash3)
	if serializeHashContent(finalHash) != buildMerkleTree(hashSlice) {
		t.Errorf("Hashes don't match: %x != %x\n", serializeHashContent(finalHash), buildMerkleTree(hashSlice))
	}
}
