package coingecko

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	bCtx "github.com/x-xyz/goapi/base/ctx"
)

func Test_CoinGecko(t *testing.T) {
	req := require.New(t)
	c := NewClient(&ClientCfg{
		HttpClient: http.Client{},
		Timeout:    10 * time.Second,
	})
	ctx := bCtx.Background()
	id := "apecoin"
	price, err := c.GetPrice(ctx, id)
	req.NoError(err)
	req.NotZero(price)
	t.Logf("%+v", price)
}

func Test_CoinGeckoHistory(t *testing.T) {
	req := require.New(t)
	c := NewClient(&ClientCfg{
		HttpClient: http.Client{},
		Timeout:    10 * time.Second,
	})
	ctx := bCtx.Background()
	id := "ethereum"
	price, err := c.GetPriceAtDate(ctx, id, "20-01-2022")
	req.NoError(err)
	req.NotZero(price)
	t.Logf("%+v", price)
}
