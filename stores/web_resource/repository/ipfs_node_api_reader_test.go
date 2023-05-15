package repository

import (
	"fmt"
	"testing"
	"time"

	ipfsapi "github.com/ipfs/go-ipfs-api"
	"github.com/stretchr/testify/require"
	bCtx "github.com/x-xyz/goapi/base/ctx"
)

func Test_ipfsNodeApiReaderRepo_Get(t *testing.T) {
	// local ipfs-node required
	if testing.Short() {
		t.Skip()
	}
	req := require.New(t)
	expectedStr := fmt.Sprintf("%s\n", `{"image":"ipfs://QmRRPWG96cmgTn2qSzjwr2qvfNEuhunv6FNeMFGa9bx6mQ","attributes":[{"trait_type":"Earring","value":"Silver Hoop"},{"trait_type":"Background","value":"Orange"},{"trait_type":"Fur","value":"Robot"},{"trait_type":"Clothes","value":"Striped Tee"},{"trait_type":"Mouth","value":"Discomfort"},{"trait_type":"Eyes","value":"X Eyes"}]}`)
	expected := []byte(expectedStr)

	ctx := bCtx.Background()
	url := "localhost:5001"
	s := ipfsapi.NewShell(url)
	timeout := 15 * time.Second
	r := NewIpfsNodeApiReaderRepo(s, timeout)
	b, err := r.Get(ctx, "QmeSjSinHpPnmXmspMjwiXyN6zS4E9zccariGR3jxcaWtq/0")
	req.NoError(err)
	req.Equal(expected, b)
}
