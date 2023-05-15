package usecase

import (
	"errors"

	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/base/ptr"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/collection"
	"github.com/x-xyz/goapi/domain/nftitem"
	"github.com/x-xyz/goapi/domain/punk"
	"github.com/x-xyz/goapi/service/query"
)

type PunkEventUseCaseCfg struct {
	Nftitem        nftitem.Repo
	CollectionRepo collection.Repo
}

type punkEventUseCase struct {
	nftitem        nftitem.Repo
	collectionRepo collection.Repo
}

func NewPunkEventUseCase(cfg *PunkEventUseCaseCfg) punk.PunkEventUseCase {
	return &punkEventUseCase{
		nftitem:        cfg.Nftitem,
		collectionRepo: cfg.CollectionRepo,
	}
}

func (u *punkEventUseCase) Transfer(ctx bCtx.Ctx, chainId domain.ChainId, event *punk.Transfer, lMeta *domain.LogMeta) error {
	ctx.WithFields(log.Fields{
		"chainId": chainId,
		"event":   event,
		"lMeta":   lMeta,
	}).Info("Transfer")

	token, err := u.nftitem.FindOne(ctx, chainId, lMeta.ContractAddress, event.TokenId)
	if err == nil {
		// nft exists
		if event.To == token.Owner {
			return nil
		}
		patchable := nftitem.PatchableNftItem{
			Owner:             &event.To,
			IndexerRetryCount: ptr.Int32(0),
		}
		if err := u.nftitem.Patch(ctx, *token.ToId(), patchable); err != nil {
			ctx.WithField("err", err).Error("nftitem.Patch failed")
			return err
		}
		return nil

	} else if !errors.Is(err, domain.ErrNotFound) {
		ctx.WithField("err", err).Error("nftitem.FindOne failed")
		return err
	}

	creator := domain.Address("")
	collection, err := u.collectionRepo.FindOne(ctx, collection.CollectionId{
		ChainId: chainId,
		Address: lMeta.ContractAddress.ToLower(),
	})
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":      err,
			"chainId":  chainId,
			"contract": lMeta.ContractAddress,
		}).Warn("failed to collectionRepo.FindOne")
	} else {
		creator = collection.Owner
	}

	// nft doesn't exists
	nft := &nftitem.NftItem{
		ChainId:         chainId,
		ContractAddress: lMeta.ContractAddress,
		TokenId:         event.TokenId,
		Owner:           event.To,
		CreatedAt:       lMeta.BlockTime,
		IsAppropriate:   ptr.Bool(true),
		ThumbnailPath:   "-",
		ImagePath:       "-",
		ImageUrl:        "https://storage.x.xyz/empty_token.jpg",
		ContentType:     "image",
		TokenType:       domain.TokenTypePunk,
		IndexerState:    nftitem.IndexerStateNew,
		Creator:         creator,
	}

	if err := u.nftitem.Create(ctx, nft); err != nil {
		if errors.Is(err, query.ErrDuplicateKey) {
			ctx.WithFields(log.Fields{
				"err":             err,
				"chainId":         chainId,
				"contractAddress": lMeta.ContractAddress,
				"tokenId":         event.TokenId,
			}).Warn("nftitem.Create failed with duplicate key")
			return nil
		}
		ctx.WithField("err", err).Error("nftitem.Create failed")
		return err
	}
	return nil
}
