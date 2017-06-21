package miner

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"github.com/lisgie/bazo_miner/protocol"
	"github.com/lisgie/bazo_miner/storage"
	"log"
	"math/big"
	"reflect"
)

//we can't use polymorphism, e.g. we can't use tx.verify() because the Transaction interface doesn't declare
//the verify method. This is because verification depends on the State (e.g., dynamic properties), which
//should only be of concern to the miner, not to the protocol package. However, this has the disadvantage
//that we have to do cas distinction here. We could use a wrapper type and do an interface for it, but
//that's too much complication
func verify(tx protocol.Transaction) bool {

	var verified bool

	switch tx.(type) {
	case *protocol.FundsTx:
		verified = verifyFundsTx(tx.(*protocol.FundsTx))
	case *protocol.AccTx:
		verified = verifyAccTx(tx.(*protocol.AccTx))
	case *protocol.ConfigTx:
		verified = verifyConfigTx(tx.(*protocol.ConfigTx))
	}
	return verified
}

func verifyFundsTx(tx *protocol.FundsTx) bool {

	var sig [24]byte
	var concatSig [64]byte
	pub1, pub2 := new(big.Int), new(big.Int)
	r, s := new(big.Int), new(big.Int)

	//fundstx only makes sense if amount > 0
	if tx.Amount == 0 || tx.Amount > protocol.MAX_MONEY {
		log.Printf("Invalid transaction amount %v\n", tx.Amount)
		return false
	}

	//check if accounts are present in the actual state
	for _, accFrom := range storage.State[tx.From] {
		accFromHash := serializeHashContent(accFrom.Address)
		for _, accTo := range storage.State[tx.To] {
			accToHash := serializeHashContent(accTo.Address)
			sig = [24]byte{}
			for cnt := 0; cnt < 24; cnt++ {
				sig[cnt] = tx.Xored[cnt] ^ accFromHash[cnt+8] ^ accToHash[cnt+8]
			}
			copy(concatSig[:24], sig[0:24])
			copy(concatSig[24:], tx.Sig[:])

			pub1.SetBytes(accFrom.Address[:32])
			pub2.SetBytes(accFrom.Address[32:])

			r.SetBytes(concatSig[:32])
			s.SetBytes(concatSig[32:])

			tx.FromHash = accFromHash
			tx.ToHash = accToHash

			txHash := tx.Hash()

			pubKey := ecdsa.PublicKey{elliptic.P256(), pub1, pub2}
			if ecdsa.Verify(&pubKey, txHash[:], r, s) == true && !reflect.DeepEqual(accFrom, accTo) {
				tx.FromHash = accFromHash
				tx.ToHash = accToHash
				return true
			}
		}
	}
	return false
}

func verifyAccTx(tx *protocol.AccTx) bool {

	r, s := new(big.Int), new(big.Int)
	pub1, pub2 := new(big.Int), new(big.Int)

	r.SetBytes(tx.Sig[:32])
	s.SetBytes(tx.Sig[32:])

	for _, rootAcc := range storage.RootKeys {
		pub1.SetBytes(rootAcc.Address[:32])
		pub2.SetBytes(rootAcc.Address[32:])

		pubKey := ecdsa.PublicKey{elliptic.P256(), pub1, pub2}
		txHash := tx.Hash()
		if ecdsa.Verify(&pubKey, txHash[:], r, s) == true {
			return true
		}
	}

	return false
}

func verifyConfigTx(tx *protocol.ConfigTx) bool {

	//account creation can only be done with a valid priv/pub key which is hard-coded
	r, s := new(big.Int), new(big.Int)
	pub1, pub2 := new(big.Int), new(big.Int)

	r.SetBytes(tx.Sig[:32])
	s.SetBytes(tx.Sig[32:])

	for _, rootAcc := range storage.RootKeys {
		pub1.SetBytes(rootAcc.Address[:32])
		pub2.SetBytes(rootAcc.Address[32:])

		pubKey := ecdsa.PublicKey{elliptic.P256(), pub1, pub2}
		txHash := tx.Hash()
		if ecdsa.Verify(&pubKey, txHash[:], r, s) == true {
			//return configBoundsChecking(tx.Id, tx.Payload)
			return true
		}
	}

	return false
}

//returns if id is in the list of possible ids and rational value for payload parameter
func parameterBoundsChecking(id uint8, payload uint64) bool {

	switch id {
	case protocol.BLOCK_SIZE_ID:
		if payload >= protocol.MIN_BLOCK_SIZE && payload <= protocol.MAX_BLOCK_SIZE {
			return true
		}
		return false
	case protocol.DIFF_INTERVAL_ID:
		if payload >= protocol.MIN_DIFF_INTERVAL && payload <= protocol.MAX_DIFF_INTERVAL {
			return true
		}
		return false
	case protocol.FEE_MINIMUM_ID:
		if payload >= protocol.MIN_FEE_MINIMUM && payload <= protocol.MAX_FEE_MINIMUM {
			return true
		}
		return false
	case protocol.BLOCK_INTERVAL_ID:
		if payload >= protocol.MIN_BLOCK_INTERVAL && payload <= protocol.MAX_BLOCK_INTERVAL {
			return true
		}
		return false
	case protocol.BLOCK_REWARD_ID:
		if payload >= protocol.MIN_BLOCK_REWARD && payload <= protocol.MAX_BLOCK_REWARD {
			return true
		}
		return false
	default:
		return false
	}
}
