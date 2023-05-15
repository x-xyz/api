package usecase

import (
	"fmt"
	"time"

	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/base/ptr"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/account"
	"github.com/x-xyz/goapi/domain/collection"
	"github.com/x-xyz/goapi/domain/erc1155"
	"github.com/x-xyz/goapi/domain/nftitem"
	"github.com/x-xyz/goapi/domain/order"
	"github.com/x-xyz/goapi/domain/token"
)

type Erc1155EventUseCaseCfg struct {
	Nftitem             nftitem.Repo
	Erc1155             erc1155.Repo
	Holding             erc1155.HoldingRepo
	CollectionRepo      collection.Repo
	ActivityHistoryRepo account.ActivityHistoryRepo
	Token               token.Usecase
	FolderUsecase       account.FolderUseCase
	OrderUseCase        order.UseCase
}

type erc1155EventUseCase struct {
	nftitem             nftitem.Repo
	erc1155             erc1155.Repo
	holding             erc1155.HoldingRepo
	activityHistoryRepo account.ActivityHistoryRepo
	collectionRepo      collection.Repo
	token               token.Usecase
	folderUsecase       account.FolderUseCase
	orderUseCase        order.UseCase
}

func NewErc1155EventUseCase(cfg *Erc1155EventUseCaseCfg) erc1155.Erc1155EventUseCase {
	return &erc1155EventUseCase{
		nftitem:             cfg.Nftitem,
		erc1155:             cfg.Erc1155,
		holding:             cfg.Holding,
		collectionRepo:      cfg.CollectionRepo,
		activityHistoryRepo: cfg.ActivityHistoryRepo,
		token:               cfg.Token,
		folderUsecase:       cfg.FolderUsecase,
		orderUseCase:        cfg.OrderUseCase,
	}
}

// Transfer do the following db update
// create nft item if not exists
// update holding of from (ignore mint, from == empty)
//   if holding become 0, delete it and decrease nftitem numOwners
// update holding of to   (ignore burn, to == empty)
//   if holding is the same as transfer value (= 0 before update), increase nftitem numOwners
// refresh active listings
func (u *erc1155EventUseCase) Transfer(ctx bCtx.Ctx, chainId domain.ChainId, transfer *erc1155.Transfer, lMeta *domain.LogMeta) error {
	ctx.WithFields(log.Fields{
		"chainId": chainId,
		"event":   transfer,
		"lMeta":   lMeta,
	}).Info("Transfer")

	nftId := nftitem.Id{ChainId: chainId, ContractAddress: lMeta.ContractAddress, TokenId: transfer.Id}

	// if dirty, update the nft item at the end
	dirty := false
	nft, err := u.getOrCreateNFTItem(ctx, nftId, lMeta.BlockTime)
	if err != nil {
		ctx.WithFields(log.Fields{
			"id":  nftId,
			"err": err,
		}).Error("ensureNftExists failed")
		return err
	}

	// update holding of from
	fromHolding := &erc1155.Holding{}
	if transfer.From != domain.EmptyAddress {
		hid := erc1155.HoldingId{ChainId: chainId, Address: lMeta.ContractAddress, TokenId: transfer.Id, Owner: transfer.From}
		_, err := u.holding.Increment(ctx, hid, -transfer.Value.Int64())
		if err != nil {
			ctx.WithField("err", err).Error("find holding of from failed")
			return err
		}
		fromHolding, err = u.holding.FindOne(ctx, hid)
		if err != nil {
			ctx.WithField("err", err).Error("find holding of from failed")
			return err
		}
		ctx.Info(fmt.Sprintf("#from holding (%s): %s: %d", fromHolding.TokenId, fromHolding.Owner, fromHolding.Balance))
		if fromHolding.Balance == 0 {
			dirty = true
			nft.NumOwners -= 1
			err = u.holding.Delete(ctx, hid)
			if err != nil {
				ctx.WithField("err", err).Error("delete holding of from failed")
				return err
			}

			err = u.folderUsecase.DeleteRelationFromAllFolders(ctx, transfer.From, nftId)
			if err != nil {
				ctx.WithField("err", err).Error("folderUsecase.DeleteRelationFromAllFolders failed")
				return err
			}
		}
	} else {
		if err := u.nftitem.IncreaseSupply(ctx, nftId, int(transfer.Value.Int64())); err != nil {
			return err
		}
	}

	// update holding of to
	toHolding := &erc1155.Holding{}
	if transfer.To != domain.EmptyAddress {
		hid := erc1155.HoldingId{ChainId: chainId, Address: lMeta.ContractAddress, TokenId: transfer.Id, Owner: transfer.To}
		_, err := u.holding.Increment(ctx, hid, transfer.Value.Int64())
		if err != nil {
			ctx.WithField("err", err).Error("find holding of to failed")
			return err
		}
		toHolding, err = u.holding.FindOne(ctx, hid)
		if err != nil {
			ctx.WithField("err", err).Error("find holding of to failed")
			return err
		}
		if toHolding.Balance == transfer.Value.Int64() {
			dirty = true
			nft.NumOwners += 1
			err = u.folderUsecase.AddNftToPublicFolder(ctx, transfer.To, nftId)
			if err != nil {
				ctx.WithField("err", err).Error("folderUsecase.AddNftToPublicFolder failed")
				return err
			}
		}
	} else {
		if err := u.nftitem.DecreaseSupply(ctx, nftId, int(transfer.Value.Int64())); err != nil {
			return err
		}
	}

	if dirty {
		patchable := nftitem.PatchableNftItem{
			NumOwners: &nft.NumOwners,
		}
		err = u.nftitem.Patch(ctx, nftId, patchable)
		if err != nil {
			ctx.WithField("err", err).Error("nftitem.Patch failed")
			return err
		}
	}

	activity := u.buildTransferActivity(chainId, transfer, lMeta)
	if err := u.activityHistoryRepo.Insert(ctx, activity); err != nil {
		ctx.WithField("err", err).Error("createTransferActivity failed")
		return err
	}

	if err := u.orderUseCase.RefreshOrders(ctx, nftId); err != nil {
		ctx.WithField("err", err).Error("orderUseCase.RefreshOrders")
		return err
	}

	if err := u.token.RefreshListingAndOfferState(ctx, nftId); err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
			"id":  nftId,
		}).Error("failed to token.RefreshListingAndOfferState")
		return err
	}

	return nil
}

func (u *erc1155EventUseCase) getOrCreateNFTItem(ctx bCtx.Ctx, id nftitem.Id, createdAt time.Time) (*nftitem.NftItem, error) {
	nft, err := u.nftitem.FindOne(ctx, id.ChainId, id.ContractAddress, id.TokenId)
	if err == nil {
		return nft, nil
	} else if err != domain.ErrNotFound {
		return nil, err
	}

	collection, err := u.collectionRepo.FindOne(ctx, collection.CollectionId{
		ChainId: id.ChainId,
		Address: id.ContractAddress.ToLower(),
	})
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":      err,
			"chainId":  id.ChainId,
			"contract": id.ContractAddress,
		}).Error("failed to collectionRepo.FindOne")
		return nil, err
	}

	nft = &nftitem.NftItem{
		ChainId:         id.ChainId,
		ContractAddress: id.ContractAddress,
		TokenId:         id.TokenId,
		IsAppropriate:   ptr.Bool(true),
		ThumbnailPath:   "-",
		ImagePath:       "-",
		ImageUrl:        "https://storage.x.xyz/empty_token.jpg",
		ContentType:     "image",
		TokenType:       domain.TokenType1155,
		IndexerState:    nftitem.IndexerStateNew,
		CreatedAt:       createdAt,
		Creator:         collection.Owner,
	}

	if err := u.nftitem.Create(ctx, nft); err != nil {
		ctx.WithField("err", err).Error("nftitem.Create failed")
		return nil, err
	}

	return nft, nil
}

func (u *erc1155EventUseCase) buildTransferActivity(chainId domain.ChainId, transfer *erc1155.Transfer, lMeta *domain.LogMeta) *account.ActivityHistory {
	activity := &account.ActivityHistory{
		ChainId:         chainId,
		ContractAddress: lMeta.ContractAddress,
		Type:            account.ActivityHistoryTypeTransfer,
		TokenId:         transfer.Id,
		Account:         transfer.From,
		To:              transfer.To,
		Quantity:        transfer.Value.String(),
		Price:           "0",
		BlockNumber:     lMeta.BlockNumber,
		TxHash:          lMeta.TxHash,
		LogIndex:        int64(lMeta.LogIndex),
		Time:            lMeta.BlockTime,
		Source:          account.SourceX,
	}

	if transfer.From == domain.EmptyAddress {
		activity.Type = account.ActivityHistoryTypeMint
	}
	return activity
}

// This use case is only for reindex
type erc1155TransferActivityUseCase struct {
	*erc1155EventUseCase
}

func NewErc1155TransferActivityUseCase(cfg *Erc1155EventUseCaseCfg) erc1155.Erc1155EventUseCase {
	return &erc1155TransferActivityUseCase{
		erc1155EventUseCase: &erc1155EventUseCase{
			nftitem:             cfg.Nftitem,
			erc1155:             cfg.Erc1155,
			holding:             cfg.Holding,
			collectionRepo:      cfg.CollectionRepo,
			activityHistoryRepo: cfg.ActivityHistoryRepo,
			token:               cfg.Token,
		},
	}
}

func (u *erc1155TransferActivityUseCase) Transfer(ctx bCtx.Ctx, chainId domain.ChainId, transfer *erc1155.Transfer, lMeta *domain.LogMeta) error {
	activity := u.buildTransferActivity(chainId, transfer, lMeta)
	return u.activityHistoryRepo.InsertTransferActivityIfNotExists(ctx, activity)
}
