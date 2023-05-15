package ethereum

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateMsgSignature(t *testing.T) {
	messageTemplate := "this is signature message template %s"
	privateKey, publicKey, err := GenerateKey()
	assert.NoError(t, err)
	address := crypto.PubkeyToAddress(*publicKey).Hex()
	nonce := "123456"
	message := []byte(fmt.Sprintf(messageTemplate, nonce))
	hash := accounts.TextHash(message)
	signature, err := crypto.Sign(hash, privateKey)
	assert.NoError(t, err)

	res, err := ValidateMsgSignature(message, hexutil.Encode(signature), address)
	assert.NoError(t, err)
	assert.True(t, res)

	// incorrect nonce
	res2, err := ValidateMsgSignature([]byte("654321"), hexutil.Encode(signature), address)
	assert.NoError(t, err)
	assert.False(t, res2)

	// incorrect signer
	_, pubKey, err := GenerateKey()
	assert.NoError(t, err)
	res3, err := ValidateMsgSignature(message, hexutil.Encode(signature), crypto.PubkeyToAddress(*pubKey).Hex())
	assert.NoError(t, err)
	assert.False(t, res3)
}

func TestValidateHashSignature(t *testing.T) {
	req := require.New(t)
	hash := hexutil.MustDecode("0x7d4a470c1f919efbc629d12c57cf5dbc7eee958d0b6d787f842944c0be83c8c3")
	sig := "0xfae5218f6165f30bf7d8798d6f1990fde8fea58c336b36c8cd3078b4d8dc2a9d0448debd2b776fb0f6bdf91d1142474d4682057d290561814172bce4641108641c"
	signer := "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
	valid, err := ValidateHashSignature(hash, sig, signer)
	req.NoError(err)
	req.True(valid)
}
