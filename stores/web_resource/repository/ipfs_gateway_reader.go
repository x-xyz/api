package repository

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain"
)

type ipfsGatewayReaderRepo struct {
	client     http.Client
	gateway    string
	ctxTimeout time.Duration
}

func NewIpfsGatewayReaderRepo(c http.Client, gateway string, timeout time.Duration) domain.WebResourceReaderRepository {
	return &ipfsGatewayReaderRepo{client: c, gateway: gateway, ctxTimeout: timeout}
}

func (r *ipfsGatewayReaderRepo) Get(c bCtx.Ctx, cid string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", r.gateway, cid)
	ctx, cancel := bCtx.WithTimeout(c, r.ctxTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := r.client.Do(req)
	if err != nil {
		ctx.WithField("cid", cid).Warn("failed with request")
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		ctx.WithFields(log.Fields{
			"cid":        cid,
			"statusCode": resp.StatusCode,
		}).Error("resp.StatusCode != 200")
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ctx.WithFields(log.Fields{
			"cid": cid,
			"err": err,
		}).Error("failed to read body")
		return nil, err
	}
	return body, nil
}
