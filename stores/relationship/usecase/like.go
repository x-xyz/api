package usecase

import (
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/like"
	"github.com/x-xyz/goapi/domain/nftitem"
)

type likeImpl struct {
	like    like.Repo
	nftitem nftitem.Repo
}

func NewLike(like like.Repo, nftitem nftitem.Repo) like.Usecase {
	return &likeImpl{like, nftitem}
}

func (im *likeImpl) GetLikers(c ctx.Ctx, chainId domain.ChainId, contract domain.Address, tokenId domain.TokenId) ([]domain.Address, error) {
	if res, err := im.like.FindAll(c, like.WithToken(chainId, contract, tokenId)); err != nil {
		c.WithField("err", err).Error("like.FinaAll failed")
		return nil, err
	} else {
		addresses := []domain.Address{}
		for _, like := range res {
			addresses = append(addresses, like.Liker)
		}
		return addresses, nil
	}
}

func (im *likeImpl) GetLikerCount(c ctx.Ctx, chainId domain.ChainId, contract domain.Address, tokenId domain.TokenId) (int, error) {
	if res, err := im.like.Count(c, like.WithToken(chainId, contract, tokenId)); err != nil {
		c.WithField("err", err).Error("like.FinaAll failed")
		return 0, err
	} else {
		return res, nil
	}
}

func (im *likeImpl) GetLikeds(c ctx.Ctx, liker domain.Address) ([]*like.Like, error) {
	if res, err := im.like.FindAll(c, like.WithLiker(liker)); err != nil {
		c.WithField("err", err).Error("like.FinaAll failed")
		return nil, err
	} else {
		return res, nil
	}
}

func (im *likeImpl) GetLikedCount(c ctx.Ctx, liker domain.Address) (int, error) {
	if res, err := im.like.Count(c, like.WithLiker(liker)); err != nil {
		c.WithField("err", err).Error("like.FinaAll failed")
		return 0, err
	} else {
		return res, nil
	}
}

func (im *likeImpl) IsLiked(c ctx.Ctx, chainId domain.ChainId, contract domain.Address, tokenId domain.TokenId, liker domain.Address) (bool, error) {
	if res, err := im.like.Count(c, like.WithToken(chainId, contract, tokenId), like.WithLiker(liker)); err != nil {
		return false, err
	} else {
		return res == 1, nil
	}
}

func (im *likeImpl) Like(c ctx.Ctx, chainId domain.ChainId, contract domain.Address, tokenId domain.TokenId, liker domain.Address) (int, error) {
	if _, err := im.nftitem.FindOne(c, chainId, contract, tokenId); err != nil {
		c.WithField("err", err).Error("nftitem.FindOne failed")
		return 0, err
	}
	if err := im.like.Create(c, like.Like{ChainId: chainId, ContractAddress: contract, TokenId: tokenId, Liker: liker}); err != nil {
		c.WithField("err", err).Error("like.Create failed")
		return 0, err
	}

	if res, err := im.nftitem.IncreaseLikeCount(c, nftitem.Id{ChainId: chainId, ContractAddress: contract, TokenId: tokenId}, 1); err != nil {
		c.WithField("err", err).Error("nftitem.IncreaseLikeCount failed")
		return 0, err
	} else {
		return int(res), nil
	}
}

func (im *likeImpl) Unlike(c ctx.Ctx, chainId domain.ChainId, contract domain.Address, tokenId domain.TokenId, liker domain.Address) (int, error) {
	if err := im.like.Delete(c, like.WithToken(chainId, contract, tokenId), like.WithLiker(liker)); err != nil {
		c.WithField("err", err).Error("like.Delete failed")
		return 0, err
	}

	if res, err := im.nftitem.IncreaseLikeCount(c, nftitem.Id{ChainId: chainId, ContractAddress: contract, TokenId: tokenId}, -1); err != nil {
		c.WithField("err", err).Error("nftitem.IncreaseLikeCount failed")
		return 0, err
	} else {
		return int(res), nil
	}
}
