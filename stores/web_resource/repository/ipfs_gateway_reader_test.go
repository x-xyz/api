package repository

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	bCtx "github.com/x-xyz/goapi/base/ctx"
)

func Test_ipfsGatewayReaderRepo_Get(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	c := http.Client{}
	timeout := 10 * time.Second

	// BAYC #0
	// ipfs://QmeSjSinHpPnmXmspMjwiXyN6zS4E9zccariGR3jxcaWtq/0
	expectedStr := fmt.Sprintf("%s\n", `{"image":"ipfs://QmRRPWG96cmgTn2qSzjwr2qvfNEuhunv6FNeMFGa9bx6mQ","attributes":[{"trait_type":"Earring","value":"Silver Hoop"},{"trait_type":"Background","value":"Orange"},{"trait_type":"Fur","value":"Robot"},{"trait_type":"Clothes","value":"Striped Tee"},{"trait_type":"Mouth","value":"Discomfort"},{"trait_type":"Eyes","value":"X Eyes"}]}`)
	expected := []byte(expectedStr)
	ctx := bCtx.Background()

	gateways := []string{
		"https://ipfs.io/ipfs",
		// "https://cloudflare-ipfs.com/ipfs", // doesn't work, even after solving captcha on browser
	}
	for _, g := range gateways {
		t.Run(g, func(t *testing.T) {
			req := require.New(t)
			r := NewIpfsGatewayReaderRepo(c, g, timeout)
			b, err := r.Get(ctx, "QmeSjSinHpPnmXmspMjwiXyN6zS4E9zccariGR3jxcaWtq/0")
			req.NoError(err)
			req.Equal(expected, b)
		})
	}
}
