package usecase

import (
	"time"

	"github.com/google/uuid"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/ip"
	"github.com/x-xyz/goapi/domain/nftitem"
)

type impl struct {
	ipRepo      ip.Repo
	nftitemRepo nftitem.Repo
}

func New(ipRepo ip.Repo, nftitemRepo nftitem.Repo) ip.UseCase {
	return &impl{ipRepo, nftitemRepo}
}

func (im *impl) Insert(ctx ctx.Ctx, listing *ip.IPListing) error {
	listing.Owner = listing.Owner.ToLower()
	if listing.IsIpOwner {
		for _, tokenId := range listing.TokenIds {
			token, err := im.nftitemRepo.FindOne(ctx, listing.ChainId, listing.ContractAddress, tokenId)
			if err != nil {
				ctx.WithFields(log.Fields{
					"err":             err,
					"chainId":         listing.ChainId,
					"contractAddress": listing.ContractAddress,
					"tokenId":         tokenId,
				}).Error("nftitemRepo.FindOne failed")
				return err
			}
			if token.Owner != listing.Owner {
				ctx.WithFields(log.Fields{
					"err":          err,
					"tokenOwner":   token.Owner,
					"listingOwner": listing.Owner,
				}).Warn("invalid listing/token owner")
				return domain.ErrBadParamInput
			}
		}
	}

	listing.ID = uuid.NewString()
	listing.CreatedAt = time.Now()
	return im.ipRepo.Insert(ctx, listing)
}

func (im *impl) FindAll(ctx ctx.Ctx, opts ...ip.FindAllOptionsFunc) ([]*ip.IPListing, error) {
	return im.ipRepo.FindAll(ctx, opts...)
}

func (im *impl) FindOne(ctx ctx.Ctx, id string) (*ip.IPListing, error) {
	return im.ipRepo.FindOne(ctx, id)
}

func (im *impl) Delete(ctx ctx.Ctx, id string) error {
	return im.ipRepo.Delete(ctx, id)
}
