package repository

import (
	"io/ioutil"
	"time"

	ipfsapi "github.com/ipfs/go-ipfs-api"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
)

type ipfsNodeApiReaderRepo struct {
	shell      *ipfsapi.Shell
	ctxTimeout time.Duration
}

func NewIpfsNodeApiReaderRepo(s *ipfsapi.Shell, timeout time.Duration) domain.WebResourceReaderRepository {
	return &ipfsNodeApiReaderRepo{shell: s, ctxTimeout: timeout}
}

func (r *ipfsNodeApiReaderRepo) Get(c ctx.Ctx, cid string) ([]byte, error) {
	ctx, cancel := ctx.WithTimeout(c, r.ctxTimeout)
	defer cancel()
	resp, err := r.shell.Request("cat", cid).Send(ctx)
	if err != nil {
		c.WithField("err", err).Error("shell.Request failed")
		return nil, err
	}
	if resp.Error != nil {
		c.WithField("resp.Error", resp.Error).Error("shell.Request failed")
		return nil, resp.Error
	}
	return ioutil.ReadAll(resp.Output)
}
