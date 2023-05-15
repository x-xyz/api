package hyype

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain/keys"
	"github.com/x-xyz/goapi/service/cache"
	"github.com/x-xyz/goapi/service/cache/provider/primitive"
)

const (
	bearerKey   = "client-id"
	apiEndpoint = "http://api.hyy.pe/api/v1"
)

type client struct {
	client  http.Client
	timeout time.Duration
	apikey  string
	cache   cache.Service
}

func NewClient(cfg *ClientCfg) Client {
	return &client{
		client:  cfg.HttpClient,
		timeout: cfg.Timeout,
		apikey:  cfg.Apikey,
		cache: cache.New(cache.ServiceConfig{
			Ttl:   time.Hour,
			Pfx:   "chainlink_cache",
			Cache: primitive.NewPrimitive("chainlink_cache", 32),
		}),
	}
}

func (c *client) GetLoresOfNft(ctx bCtx.Ctx, contract, tokenId string, skip, limit int) ([]byte, error) {
	url := fmt.Sprintf("%s/lores?searchKey=%s-%s&skip=%d&limit=%d", apiEndpoint, contract, tokenId, skip, limit)

	var data []byte
	key := keys.RedisKey(fmt.Sprintf("%s-%s", contract, tokenId))
	if err := c.cache.GetByFunc(ctx, key, &data, func() (interface{}, error) {
		data, err := c.get(ctx, url)
		if err != nil {
			ctx.WithFields(log.Fields{
				"url": url,
				"err": err,
			}).Error("c.get failed")
			return nil, err
		}
		return &data, err
	}); err != nil {
		ctx.WithFields(log.Fields{
			"url": url,
			"err": err,
		}).Error("cache.GetByFunc failed")
		return nil, err
	}

	return data, nil
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
