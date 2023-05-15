package ethereum

import (
	"bytes"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

func ValidateMsgSignature(message []byte, signature, signer string) (bool, error) {
	return validateSignature(message, signature, signer, true)
}

func ValidateHashSignature(hash []byte, signature, signer string) (bool, error) {
	return validateSignature(hash, signature, signer, false)
}

func validateSignature(data []byte, signature, signer string, applyTextHash bool) (bool, error) {
	hash := data
	if applyTextHash {
		hash = accounts.TextHash(data)
	}
	address := common.HexToAddress(signer)
	sig := hexutil.MustDecode(signature)
	recoveredAddress, err := ecRecover(hash, sig)
	if err != nil {
		return false, err
	}
	return bytes.Equal(address.Bytes(), recoveredAddress.Bytes()), nil
}

// ecRecover returns the address for the account that was used to create the signature.
// copy of internal go-ethereum function:
// https://github.com/ethereum/go-ethereum/blob/v1.10.9/internal/ethapi/api.go#L524
func ecRecover(data []byte, sig []byte) (common.Address, error) {
	if len(sig) != crypto.SignatureLength {
		return common.Address{}, fmt.Errorf("signature must be %d bytes long", crypto.SignatureLength)
	}

	// support both versions of `eth_sign` responses
	//	@see	https://github.com/ethereumjs/ethereumjs-util/blob/master/src/signature.ts#L112
	if sig[crypto.RecoveryIDOffset] < 27 {
		sig[crypto.RecoveryIDOffset] += 27
	}

	if sig[crypto.RecoveryIDOffset] != 27 && sig[crypto.RecoveryIDOffset] != 28 {
		return common.Address{}, fmt.Errorf("invalid Ethereum signature (V is not 27 or 28)")
	}

	sig[crypto.RecoveryIDOffset] -= 27 // Transform yellow paper V from 27/28 to 0/1

	rpk, err := crypto.SigToPub(data, sig)

	if err != nil {
		return common.Address{}, err
	}

	return crypto.PubkeyToAddress(*rpk), nil
}
