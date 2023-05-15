package repository

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	bCtx "github.com/x-xyz/goapi/base/ctx"
)

func Test_arReaderRepo_Get(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	req := require.New(t)
	c := http.Client{}
	timeout := 10 * time.Second

	// White Rabbit 0x97ed92e744c10fdd5d403a756239c4069e415e79 #1
	url := "ar://eXcwlbsV1BiRGCsGKXa60Mj0i-xDZU0k95l_ysNwv_w/1.json"
	expectedStr := `{
  "name": "White Rabbit Producer Pass Chapter 1",
  "collection": "White Rabbit",
  "description": "These are utility NFTs for voting on shibuya.xyz at the end of each chapter to earn $WRAB, which represents fractionalized NFT ownership of the final film.",
  "animation_url": "https://arweave.net/3q0DQXGP_i_JIZ9L4JKeerpF3IKj_Re1b1dkm8Y8RlE",
  "image": "https://arweave.net/IjMqj5IFKl9iP2c1BoZ81-__o5cu1QI_V_9E-3SUsuU",
  "external_url": "www.shibuya.xyz",
  "attributes": [
    {
      "trait_type": "Artist",
      "value": "Maciej Kuciara"
    },
    {
      "trait_type": "Artist",
      "value": "Pplpleasr"
    },
    {
      "trait_type": "Chapter",
      "value": "1"
    },
    {
      "trait_type": "Film Name",
      "value": "White Rabbit"
    },
    {
      "trait_type": "Fractionalized Token",
      "value": "$WRAB"
    },
    {
      "trait_type": "Place",
      "value": "Shibuya.xyz"
    },
    {
      "trait_type": "Devs",
      "value": "PleasrDevs"
    }
  ]
}
`
	expected := []byte(expectedStr)
	ctx := bCtx.Background()

	r := NewArReaderRepo(c, timeout, nil)
	b, err := r.Get(ctx, url)
	req.NoError(err)
	req.Equal(expected, b)
}
