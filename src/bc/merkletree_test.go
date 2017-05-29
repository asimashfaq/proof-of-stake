package bc

import (
	"testing"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
)

func TestBuildMerkleTree(t *testing.T) {

	var funds []fundsTx
	var hash1,hash2,hash3 [32]byte
	var tmpHash []byte
	var tx, tx2, tx3 fundsTx

	//generating a private key and prepare data
	privA,_ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	tx,_ = ConstrFundsTx(0x01,23,1,0,[32]byte{'0'},[32]byte{'1'},privA)
	tx2,_ = ConstrFundsTx(0x01,26,1,1,[32]byte{'0'},[32]byte{'1'},privA)
	tx3,_ = ConstrFundsTx(0x02,25,1,34,[32]byte{'2'},[32]byte{'5'}, privA)

	//testing with 1,2,3 nodes
	toHash1 := struct{
		Header byte
		Amount [8]byte
		Fee [8]byte
		TxCnt [3]byte
		From [32]byte
		To [32]byte
	}{
		tx.Header,
		tx.Amount,
		tx.Fee,
		tx.TxCnt,
		tx.fromHash,
		tx.toHash,
	}
	hash1 = serializeHashContent(toHash1)

	toHash2 := struct{
		Header byte
		Amount [8]byte
		Fee [8]byte
		TxCnt [3]byte
		From [32]byte
		To [32]byte
	}{
		tx2.Header,
		tx2.Amount,
		tx2.Fee,
		tx2.TxCnt,
		tx2.fromHash,
		tx2.toHash,
	}
	hash2 = serializeHashContent(toHash2)

	toHash3 := struct{
		Header byte
		Amount [8]byte
		Fee [8]byte
		TxCnt [3]byte
		From [32]byte
		To [32]byte
	}{
		tx3.Header,
		tx3.Amount,
		tx3.Fee,
		tx3.TxCnt,
		tx3.fromHash,
		tx3.toHash,
	}
	hash3 = serializeHashContent(toHash3)

	//test with one node
	//self hash
	funds = append(funds,tx)
	tmpHash = append(hash1[:],hash1[:]...)
	if serializeHashContent(tmpHash) != buildMerkleTree(funds) {
		t.Errorf("Hashes don't match: %x != %x\n", serializeHashContent(tmpHash), buildMerkleTree(funds))
	}

	//two nodes
	funds = append(funds,tx2)
	tmpHash = append(hash1[:],hash2[:]...)
	if serializeHashContent(tmpHash) != buildMerkleTree(funds) {
		t.Errorf("Hashes don't match: %x != %x\n", serializeHashContent(tmpHash), buildMerkleTree(funds))
	}

	//three nodes
	funds = append(funds,tx3)
	tmpHash = append(hash1[:],hash2[:]...)
	tmpHashHash := serializeHashContent(tmpHash)
	tmpHash2 := append(hash3[:],hash3[:]...)
	tmpHashHash2 := serializeHashContent(tmpHash2)
	finalHash := append(tmpHashHash[:],tmpHashHash2[:]...)
	if serializeHashContent(finalHash) != buildMerkleTree(funds) {
		t.Errorf("Hashes don't match: %x != %x\n", serializeHashContent(finalHash), buildMerkleTree(funds))
	}
}
