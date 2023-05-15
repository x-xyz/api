package opensea

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/x-xyz/goapi/domain"

	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
)

const (
	bearerKey = "X-API-KEY"
	v1Api     = "https://api.opensea.io/api/v1"
)

func NewClient(cfg *ClientCfg) Client {
	return &client{
		client:  cfg.HttpClient,
		timeout: cfg.Timeout,
		apikey:  cfg.Apikey,
	}
}

type client struct {
	client  http.Client
	timeout time.Duration
	apikey  string
}

func (c *client) GetAssetContractByAddress(ctx bCtx.Ctx, addr string) (*AssetContractResp, error) {
	url := fmt.Sprintf("%s/asset_contract/%s", v1Api, addr)
	data, err := c.get(ctx, url)
	if err != nil {
		ctx.WithFields(log.Fields{
			"url": url,
			"err": err,
		}).Error("c.get failed")
		return nil, err
	}
	resp := &AssetContractResp{}
	if err := json.Unmarshal(data, resp); err != nil {
		ctx.WithField("err", err).Error("json.Unmarshal failed")
		return nil, err
	}
	return resp, nil
}

func (c *client) GetCollectionBySlug(ctx bCtx.Ctx, slug string) (*CollectionResp, error) {
	url := fmt.Sprintf("%s/collection/%s", v1Api, slug)
	data, err := c.get(ctx, url)
	if err != nil {
		ctx.WithFields(log.Fields{
			"url": url,
			"err": err,
		}).Error("c.get failed")
		return nil, err
	}
	resp := &CollectionResp{}
	if err := json.Unmarshal(data, resp); err != nil {
		ctx.WithField("err", err).Error("json.Unmarshal failed")
		return nil, err
	}
	return resp, nil
}

func (c *client) get(ctx bCtx.Ctx, url string) ([]byte, error) {
	ctx, cancel := bCtx.WithTimeout(ctx, c.timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		ctx.WithFields(log.Fields{
			"url": url,
			"err": err,
		}).Error("NewRequestWithContext failed")
		return nil, err
	}
	req.Header.Set(bearerKey, c.apikey)
	resp, err := c.client.Do(req)
	if err != nil {
		ctx.WithFields(log.Fields{
			"url": url,
			"err": err,
		}).Error("client.Do failed")
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		ctx.WithFields(log.Fields{
			"url":        url,
			"statusCode": resp.StatusCode,
		}).Error("resp.StatusCode != 200")
		return nil, ErrStatusCodeNotOk
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ctx.WithFields(log.Fields{
			"url": url,
			"err": err,
		}).Error("failed to read body")
		return nil, err
	}
	return body, nil
}

func (c *client) GetEvent(ctx bCtx.Ctx, opts ...GetEventOptionsFunc) (*EventResp, error) {
	opt, err := ParseGetEventOptions(opts...)
	if err != nil {
		return nil, err
	}

	base, err := url.Parse(fmt.Sprintf("%s/events", v1Api))
	if err != nil {
		return nil, err
	}

	params := url.Values{}

	if opt.ContractAddress != nil {
		params.Add("asset_contract_address", opt.ContractAddress.ToLowerStr())
	}

	if opt.EventType != nil {
		params.Add("event_type", string(*opt.EventType))
	}

	if opt.Before != nil {
		params.Add("occurred_before", opt.Before.Format(time.RFC3339))
	}

	if opt.After != nil {
		params.Add("occurred_after", opt.After.Format(time.RFC3339))
	}

	if opt.Cursor != nil {
		params.Add("cursor", *opt.Cursor)
	}

	base.RawQuery = params.Encode()
	url := base.String()

	data, err := c.get(ctx, url)
	if err != nil {
		ctx.WithFields(log.Fields{
			"url": url,
			"err": err,
		}).Error("c.get failed")
		return nil, err
	}

	resp := EventResp{}
	if err := json.Unmarshal(data, &resp); err != nil {
		ctx.WithField("err", err).Error("json.Unmarshal failed")
		return nil, err
	}

	return &resp, nil
}

func (c *client) GetAsset(ctx bCtx.Ctx, collectionSlug string, tokenId string) (*AssetsResp, error) {
	base, err := url.Parse(fmt.Sprintf("%s/assets", v1Api))
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Add("order_direction", "desc")
	params.Add("limit", "50")
	params.Add("include_orders", "true")
	params.Add("token_ids", tokenId)
	params.Add("collection_slug", collectionSlug)

	base.RawQuery = params.Encode()
	url := base.String()

	data, err := c.get(ctx, url)
	if err != nil {
		ctx.WithFields(log.Fields{
			"url": url,
			"err": err,
		}).Error("c.get failed")
		return nil, err
	}

	resp := AssetsResp{}

	if err := json.Unmarshal(data, &resp); err != nil {
		ctx.WithField("err", err).Error("json.Unmarshal failed")
		return nil, err
	}

	return &resp, nil
}

func (c *client) GetAssetByOwner(ctx bCtx.Ctx, owner domain.Address, cursor string) (*AssetsResp, error) {
	base, err := url.Parse(fmt.Sprintf("%s/assets", v1Api))
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Add("owner", owner.ToLowerStr())
	params.Add("include_orders", "true")
	params.Add("limit", "50")
	if cursor != "" {
		params.Add("cursor", cursor)
	}
	base.RawQuery = params.Encode()
	url := base.String()

	data, err := c.get(ctx, url)
	if err != nil {
		ctx.WithFields(log.Fields{
			"url": url,
			"err": err,
		}).Error("c.get failed")
		return nil, err
	}

	resp := AssetsResp{}

	if err := json.Unmarshal(data, &resp); err != nil {
		ctx.WithField("err", err).Error("json.Unmarshal failed")
		return nil, err
	}

	return &resp, nil
}
