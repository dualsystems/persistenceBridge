package casp

import (
	"crypto/ecdsa"
	"encoding/hex"
	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/persistenceOne/persistenceBridge/utilities/logging"
	"math/big"
)

// Should include prefix "04"
func GetTMPubKey(caspPubKey string) cryptotypes.PubKey {
	x, y := getXY(caspPubKey)

	pubKey := ecdsa.PublicKey{
		Curve: btcec.S256(),
		X:     &x,
		Y:     &y,
	}
	pubkeyObject := (*btcec.PublicKey)(&pubKey)
	pk := pubkeyObject.SerializeCompressed()
	return &secp256k1.PubKey{Key: pk}
}

// Should include prefix "04"
func GetEthPubKey(caspPubKey string) ecdsa.PublicKey {
	x, y := getXY(caspPubKey)
	publicKey := ecdsa.PublicKey{
		Curve: crypto.S256(),
		X:     &x,
		Y:     &y,
	}
	return publicKey
}

// Should include prefix "04"
func getXY(caspPubKey string) (big.Int, big.Int) {
	pubKeyBytes, err := hex.DecodeString(string([]rune(caspPubKey)[2:])) // uncompressed pubkey
	if err != nil {
		logging.Fatal(err)
	}
	var x big.Int
	x.SetBytes(pubKeyBytes[0:32])
	var y big.Int
	y.SetBytes(pubKeyBytes[32:])
	return x, y
}
