package opensea

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
)

func Test_Opensea(t *testing.T) {
	req := require.New(t)
	c := NewClient(&ClientCfg{
		HttpClient: http.Client{},
		Timeout:    10 * time.Second,
		Apikey:     "api_key",
	})
	bayc := "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d"
	ctx := bCtx.Background()
	resp, err := c.GetAssetContractByAddress(ctx, bayc)
	req.NoError(err)
	req.Equal("boredapeyachtclub", resp.Collection.Slug)
	t.Logf("%+v", resp)

	cResp, err := c.GetCollectionBySlug(ctx, resp.Collection.Slug)
	req.NoError(err)
	req.Equal("Bored Ape Yacht Club", cResp.Collection.Name)
	t.Logf("%+v", cResp)

	eventResp, err := c.GetEvent(
		ctx,
		WithContractAddress(domain.Address(bayc)),
		WithEventType(EventTypeSuccessful),
		WithBefore(time.Date(2022, 4, 5, 23, 99, 99, 0, time.UTC)),
		WithAfter(time.Date(2022, 4, 5, 0, 0, 0, 0, time.UTC)),
		WithCursor(""),
	)
	req.NoError(err)
	req.Equal(eventResp.Next, "")
	req.NotEqual(len(eventResp.AssetEvents), 0)
	t.Logf("%+v", eventResp)

	assetResp, err := c.GetAsset(
		ctx,
		"boredapeyachtclub",
		"9719",
	)
	req.NoError(err)
	t.Logf("%+v\n", assetResp)
}
