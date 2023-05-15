package coingecko

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/shopspring/decimal"
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain/keys"
	"github.com/x-xyz/goapi/service/cache"
	"github.com/x-xyz/goapi/service/cache/provider/primitive"
)

const api = "https://api.coingecko.com/api/v3"

func NewClient(cfg *ClientCfg) Client {
	return &client{
		client:  cfg.HttpClient,
		timeout: cfg.Timeout,
		cache: cache.New(cache.ServiceConfig{
			Ttl:   time.Minute,
			Pfx:   "coingecko_cache",
			Cache: primitive.NewPrimitive("coingecko_cache", 4),
		}),
	}
}

type client struct {
	client  http.Client
	timeout time.Duration
	cache   cache.Service
}

func (c *client) GetPrice(ctx bCtx.Ctx, id string) (decimal.Decimal, error) {
	key := keys.RedisKey(id)
	var price decimal.Decimal
	if err := c.cache.GetByFunc(ctx, key, &price, func() (interface{}, error) {
		if res, err := c.getPrice(ctx, id); err != nil {
			return &decimal.Zero, err
		} else {
			return res, nil
		}
	}); err != nil {
		return decimal.Zero, err
	}
	return price, nil
}

func (c *client) getPrice(ctx bCtx.Ctx, id string) (*decimal.Decimal, error) {
	params := url.Values{
		"vs_currency": {"usd"},
		"ids":         {id},
	}
	url := fmt.Sprintf("%s/coins/markets/?%s", api, params.Encode())
	data, err := c.get(ctx, url)
	if err != nil {
		ctx.WithFields(log.Fields{
			"url": url,
			"err": err,
		}).Error("c.get failed")
		return &decimal.Zero, err
	}
	resp := &Markets{}
	if err := json.Unmarshal(data, resp); err != nil {
		ctx.WithField("err", err).Error("json.Unmarshal failed")
		return &decimal.Zero, err
	}
	if len(*resp) != 1 {
		ctx.Error(ErrMarketsLen)
		return &decimal.Zero, ErrMarketsLen
	}
	price := decimal.NewFromFloat((*resp)[0].CurrentPrice)
	return &price, nil
}

func (c *client) GetPriceAtDate(ctx bCtx.Ctx, id string, date string) (decimal.Decimal, error) {
	key := keys.RedisKey("history", id, date)
	var price decimal.Decimal
	if err := c.cache.GetByFunc(ctx, key, &price, func() (interface{}, error) {
		if res, err := c.getPriceAtDate(ctx, id, date); err != nil {
			return &decimal.Zero, err
		} else {
			return &res, nil
		}
	}); err != nil {
		return decimal.Zero, err
	}
	return price, nil
}

func (c *client) getPriceAtDate(ctx bCtx.Ctx, id string, date string) (decimal.Decimal, error) {
	params := url.Values{
		"date": {date},
	}

	p := fmt.Sprintf("%s/coins/%s/history", api, id)

	url, err := url.Parse(p)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":  err,
			"path": p,
		}).Error("failed to parse url")
		return decimal.Zero, err
	}

	url.RawQuery = params.Encode()

	data, err := c.get(ctx, url.String())
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
			"url": url.String(),
		}).Error("failed to c.get")
		return decimal.Zero, err
	}

	resp := History{}

	if err := json.Unmarshal(data, &resp); err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("json.Unmarshal failed")
		return decimal.Zero, err
	}

	price := decimal.NewFromFloat(resp.MarketData.CurrentPrice.Usd)

	return price, nil
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
