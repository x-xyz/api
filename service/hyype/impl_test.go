package hyype

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	bCtx "github.com/x-xyz/goapi/base/ctx"
)

func Test_Hyype(t *testing.T) {
	req := require.New(t)
	c := NewClient(&ClientCfg{
		HttpClient: http.Client{},
		Timeout:    10 * time.Second,
		Apikey:     "api_key",
	})
	contract := "0x18c7766a10df15df8c971f6e8c1d2bba7c7a410b"
	tokenId := "4858"
	ctx := bCtx.Background()
	resp, err := c.GetLoresOfNft(ctx, contract, tokenId, 0, 10)
	req.NoError(err)
	t.Logf("%+v", string(resp))
}
