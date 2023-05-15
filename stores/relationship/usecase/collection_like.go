package usecase

import (
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/collection"
	"github.com/x-xyz/goapi/domain/like"
)

type collectionLikeImpl struct {
	like       like.CollectionLikeRepo
	collection collection.Repo
}

func NewCollectionLike(like like.CollectionLikeRepo, collection collection.Repo) like.CollectionLikeUsecase {
	return &collectionLikeImpl{like, collection}
}

func (im *collectionLikeImpl) GetLikers(c ctx.Ctx, id collection.CollectionId) ([]domain.Address, error) {
	if res, err := im.like.FindAll(c, like.WithChainId(id.ChainId), like.WithContract(id.Address)); err != nil {
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

func (im *collectionLikeImpl) GetLikerCount(c ctx.Ctx, id collection.CollectionId) (int, error) {
	if res, err := im.like.Count(c, like.WithChainId(id.ChainId), like.WithContract(id.Address)); err != nil {
		c.WithField("err", err).Error("like.FinaAll failed")
		return 0, err
	} else {
		return res, nil
	}
}

func (im *collectionLikeImpl) GetLikeds(c ctx.Ctx, liker domain.Address) ([]*like.CollectionLike, error) {
	if res, err := im.like.FindAll(c, like.WithLiker(liker)); err != nil {
		c.WithField("err", err).Error("like.FinaAll failed")
		return nil, err
	} else {
		return res, nil
	}
}

func (im *collectionLikeImpl) GetLikedCount(c ctx.Ctx, liker domain.Address) (int, error) {
	if res, err := im.like.Count(c, like.WithLiker(liker)); err != nil {
		c.WithField("err", err).Error("like.FinaAll failed")
		return 0, err
	} else {
		return res, nil
	}
}

func (im *collectionLikeImpl) IsLiked(c ctx.Ctx, id collection.CollectionId, liker domain.Address) (bool, error) {
	if res, err := im.like.Count(c, like.WithChainId(id.ChainId), like.WithContract(id.Address), like.WithLiker(liker)); err != nil {
		return false, err
	} else {
		return res == 1, nil
	}
}

func (im *collectionLikeImpl) Like(c ctx.Ctx, id collection.CollectionId, liker domain.Address) (int, error) {
	if _, err := im.collection.FindOne(c, id); err != nil {
		if err != domain.ErrNotFound {
			c.WithField("err", err).Error("collection.FindOne failed")
		}
		return 0, err
	}
	if err := im.like.Create(c, like.CollectionLike{ChainId: id.ChainId, ContractAddress: id.Address, Liker: liker}); err != nil {
		c.WithField("err", err).Error("like.Create failed")
		return 0, err
	}

	if res, err := im.collection.IncreaseLikeCount(c, id, 1); err != nil {
		c.WithField("err", err).Error("collection.IncreaseLikeCount failed")
		return 0, err
	} else {
		return int(res), nil
	}
}

func (im *collectionLikeImpl) Unlike(c ctx.Ctx, id collection.CollectionId, liker domain.Address) (int, error) {
	if err := im.like.Delete(c, like.WithChainId(id.ChainId), like.WithContract(id.Address), like.WithLiker(liker)); err != nil {
		c.WithField("err", err).Error("like.Delete failed")
		return 0, err
	}

	if res, err := im.collection.IncreaseLikeCount(c, id, -1); err != nil {
		c.WithField("err", err).Error("collection.IncreaseLikeCount failed")
		return 0, err
	} else {
		return int(res), nil
	}
}
