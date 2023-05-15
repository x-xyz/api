package repository

import (
	"io/ioutil"
	"net/http"
	"time"

	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain"
	"golang.org/x/xerrors"
)

type httpReaderRepo struct {
	client     http.Client
	ctxTimeout time.Duration
	headers    map[string]string
}

func NewHttpReaderRepo(client http.Client, timeout time.Duration, headers map[string]string) domain.WebResourceReaderRepository {
	return &httpReaderRepo{client: client, ctxTimeout: timeout, headers: headers}
}

func (r *httpReaderRepo) Get(c bCtx.Ctx, url string) ([]byte, error) {
	ctx, cancel := bCtx.WithTimeout(c, r.ctxTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	if r.headers != nil {
		for k, v := range r.headers {
			req.Header.Set(k, v)
		}
	}
	resp, err := r.client.Do(req)
	if err != nil {
		ctx.WithField("url", url).Warn("failed with request")
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		ctx.WithFields(log.Fields{
			"url":        url,
			"statusCode": resp.StatusCode,
		}).Error("resp.StatusCode != 200")
		return nil, xerrors.Errorf("resp.StatusCode != 200")
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
