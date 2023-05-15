package ethereum

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/crypto"
)

func GenerateKey() (*ecdsa.PrivateKey, *ecdsa.PublicKey, error) {
	if privateKey, err := crypto.GenerateKey(); err != nil {
		return nil, nil, err
	} else {
		publicKey := privateKey.Public().(*ecdsa.PublicKey)
		return privateKey, publicKey, nil
	}
}
