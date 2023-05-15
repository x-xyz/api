package usecase

import (
	"math/big"

	"github.com/shopspring/decimal"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/service/chainlink"
)

type impl struct {
	chainlink chainlink.Chainlink
	paytoken  domain.PayTokenRepo
}

func New(
	chainlink chainlink.Chainlink,
	paytoken domain.PayTokenRepo,
) domain.ChainlinkUsacase {
	return &impl{chainlink: chainlink, paytoken: paytoken}
}

func (im *impl) GetLatestAnswer(c ctx.Ctx, chainId domain.ChainId, tokenAddr domain.Address) (decimal.Decimal, error) {
	paytoken, err := im.paytoken.FindOne(c, chainId, tokenAddr)
	if err != nil {
		c.WithFields(log.Fields{
			"err":          err,
			"chainId":      chainId,
			"tokenAddress": tokenAddr,
		}).Error("paytoken.FindOne failed")
		return decimal.NewFromInt(0), err
	}

	if len(paytoken.ChainlinkProxyAddress) == 0 {
		return decimal.Zero, domain.ErrNoPriceFeed
	}

	rawVal, err := im.chainlink.GetLatestAnswer(c, chainId, paytoken.ChainlinkProxyAddress)
	if err != nil {
		c.WithFields(log.Fields{
			"err":          err,
			"chainId":      chainId,
			"tokenAddress": tokenAddr,
		}).Error("chainlink.GetLatestAnswer failed")
		return decimal.NewFromInt(0), err
	}

	return decimal.NewFromBigInt(rawVal, -paytoken.Decimals), nil
}

func (im *impl) GetLatestAnswerAt(c ctx.Ctx, chainId domain.ChainId, tokenAddr domain.Address, blk uint64) (decimal.Decimal, error) {
	paytoken, err := im.paytoken.FindOne(c, chainId, tokenAddr)
	if err != nil {
		c.WithFields(log.Fields{
			"err":          err,
			"chainId":      chainId,
			"tokenAddress": tokenAddr,
			"blk":          blk,
		}).Error("paytoken.FindOne failed")
		return decimal.NewFromInt(0), err
	}

	if len(paytoken.ChainlinkProxyAddress) == 0 {
		return decimal.Zero, domain.ErrNoPriceFeed
	}

	rawVal, err := im.chainlink.GetLatestAnswerAt(c, chainId, paytoken.ChainlinkProxyAddress, new(big.Int).SetUint64(blk))
	if err != nil {
		c.WithFields(log.Fields{
			"err":          err,
			"chainId":      chainId,
			"tokenAddress": tokenAddr,
			"blk":          blk,
		}).Error("chainlink.GetLatestAnswerAt failed")
		return decimal.NewFromInt(0), err
	}

	return decimal.NewFromBigInt(rawVal, -paytoken.Decimals), nil
}
