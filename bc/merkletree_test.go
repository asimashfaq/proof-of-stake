package bc

import (
	"testing"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
)

func TestBuildMerkleTree(t *testing.T) {

	var hashSlice [][32]byte
	var hashSlice2 [][32]byte
	var hashSlice3 [][32]byte
	var hash1,hash2,hash3 [32]byte
	var tmpHash []byte
	var tx *fundsTx
	var tx2 *accTx
	var tx3 *configTx

	//generating a private key and prepare data
	privA,_ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	tx,_ = ConstrFundsTx(0x01,23,1,0,[32]byte{'0'},[32]byte{'1'},privA)
	tx2,_ = ConstrAccTx(23,privA)
	tx3,_ = ConstrConfigTx(0x02,2,5000,34,privA)

	//testing with 1,2,3 nodes
	hash1 = hashFundsTx(tx)
	hash2 = hashAccTx(tx2)
	hash3 = hashConfigTx(tx3)

	//test with one node
	//self hash
	tmpHash = append(hash1[:],hash1[:]...)
	hashSlice = append(hashSlice,hash1)
	if serializeHashContent(tmpHash) != buildMerkleTree(hashSlice) {
		t.Errorf("Hashes don't match: %x != %x\n", serializeHashContent(tmpHash), buildMerkleTree(hashSlice))
	}

	//two nodes
	tmpHash = append(hash1[:],hash2[:]...)
	hashSlice2 = append(hashSlice2, hash2)
	if serializeHashContent(tmpHash) != buildMerkleTree(hashSlice, hashSlice2) {
		t.Errorf("Hashes don't match: %x != %x\n", serializeHashContent(tmpHash), buildMerkleTree(hashSlice))
	}

	//three nodes
	tmpHash = append(hash1[:],hash2[:]...)
	tmpHashHash := serializeHashContent(tmpHash)
	tmpHash2 := append(hash3[:],hash3[:]...)
	tmpHashHash2 := serializeHashContent(tmpHash2)
	finalHash := append(tmpHashHash[:],tmpHashHash2[:]...)
	hashSlice3 = append(hashSlice3,hash3)
	if serializeHashContent(finalHash) != buildMerkleTree(hashSlice,hashSlice2,hashSlice3) {
		t.Errorf("Hashes don't match: %x != %x\n", serializeHashContent(finalHash), buildMerkleTree(hashSlice,hashSlice2,hashSlice3))
	}
}
