package usecase

import (
	"errors"

	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/base/ptr"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/account"
	"github.com/x-xyz/goapi/domain/collection"
	"github.com/x-xyz/goapi/domain/erc721/contract"
	"github.com/x-xyz/goapi/domain/nftitem"
	"github.com/x-xyz/goapi/domain/order"
	"github.com/x-xyz/goapi/domain/token"
	"github.com/x-xyz/goapi/service/query"
)

type Erc721EventUseCaseCfg struct {
	Nftitem                   nftitem.Repo
	Erc721                    contract.Repo
	CollectionRepo            collection.Repo
	Token                     token.Usecase
	FolderRepo                account.FolderRepo
	FolderNftRelationshipRepo account.FolderNftRelationshipRepo
	ActivityHistoryRepo       account.ActivityHistoryRepo
	FolderUsecase             account.FolderUseCase
	OrderUseCase              order.UseCase
}

type erc721EventUseCase struct {
	nftitem                   nftitem.Repo
	erc721                    contract.Repo
	token                     token.Usecase
	collectionRepo            collection.Repo
	folderRepo                account.FolderRepo
	folderNftRelationshipRepo account.FolderNftRelationshipRepo
	activityHistoryRepo       account.ActivityHistoryRepo
	folderUsecase             account.FolderUseCase
	orderUseCase              order.UseCase
}

func NewErc721EventUseCase(cfg *Erc721EventUseCaseCfg) contract.Erc721EventUseCase {
	return &erc721EventUseCase{
		nftitem:                   cfg.Nftitem,
		erc721:                    cfg.Erc721,
		token:                     cfg.Token,
		collectionRepo:            cfg.CollectionRepo,
		folderRepo:                cfg.FolderRepo,
		folderNftRelationshipRepo: cfg.FolderNftRelationshipRepo,
		activityHistoryRepo:       cfg.ActivityHistoryRepo,
		folderUsecase:             cfg.FolderUsecase,
		orderUseCase:              cfg.OrderUseCase,
	}
}
func (u *erc721EventUseCase) Transfer(ctx bCtx.Ctx, chainId domain.ChainId, event *contract.TransferEvent, lMeta *domain.LogMeta) error {
	ctx.WithFields(log.Fields{
		"chainId": chainId,
		"event":   event,
		"lMeta":   lMeta,
	}).Info("Transfer")

	id := nftitem.Id{ChainId: chainId, ContractAddress: lMeta.ContractAddress, TokenId: event.TokenId}
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

		if err := u.orderUseCase.RefreshOrders(ctx, id); err != nil {
			ctx.WithFields(log.Fields{
				"err": err,
				"id":  id,
			}).Error("failed to orderUseCase.RefreshOrders")
			return err
		}

		if err := u.token.RefreshListingAndOfferState(ctx, id); err != nil {
			ctx.WithFields(log.Fields{
				"err": err,
				"id":  id,
			}).Error("failed to token.RefreshListingAndOfferState")
			return err
		}

		if err := u.moveNftToNewOwnerPublicFolder(ctx, event.To, id); err != nil {
			ctx.WithFields(log.Fields{
				"id":  id,
				"err": err,
			}).Error("moveNftToNewOwnerPublicFolder failed")
			return err
		}

		activity := u.buildTransferActivity(chainId, event, lMeta)
		if err := u.activityHistoryRepo.Insert(ctx, activity); err != nil {
			ctx.WithFields(log.Fields{
				"id":  id,
				"err": err,
			}).Error("createTransferActivity failed")
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
		TokenType:       domain.TokenType721,
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
	if err := u.moveNftToNewOwnerPublicFolder(ctx, event.To, id); err != nil {
		ctx.WithFields(log.Fields{
			"id":  id,
			"err": err,
		}).Error("moveNftToNewOwnerPublicFolder failed")
		return err
	}
	activity := u.buildTransferActivity(chainId, event, lMeta)
	if err := u.activityHistoryRepo.Insert(ctx, activity); err != nil {
		ctx.WithFields(log.Fields{
			"activity": activity,
			"id":       id,
			"err":      err,
		}).Error("createTransferActivity failed")
		return err
	}
	return nil
}

func (u *erc721EventUseCase) moveNftToNewOwnerPublicFolder(ctx bCtx.Ctx, owner domain.Address, id nftitem.Id) error {
	if err := u.folderNftRelationshipRepo.DeleteAllRelationsByNftitem(ctx, id); err != nil {
		ctx.WithFields(log.Fields{
			"id":  id,
			"err": err,
		}).Error("DeleteAllRelationsByNftitem failed")
		return err
	}
	folders, err := u.folderRepo.GetFolders(ctx,
		account.WithBuiltIn(true),
		account.WithOwner(owner),
		account.WithPrivate(false),
	)
	if err != nil {
		ctx.WithFields(log.Fields{
			"builtIn": true,
			"owner":   owner,
			"private": false,
			"err":     err,
		}).Error("folderRepo.GetFolders failed")
		return err
	}
	if len(folders) != 1 {
		// no public folder, skip it
		return nil
	}
	if err := u.folderNftRelationshipRepo.AddNftitemsToFolder(ctx, []nftitem.Id{id}, folders[0].Id); err != nil {
		ctx.WithFields(log.Fields{
			"nftId":    id,
			"folderId": folders[0].Id,
			"err":      err,
		}).Error("folderNftRelationshipRepo.AddNftitemsToFolder failed")
		return err
	}

	if err := u.folderUsecase.RefreshCount(ctx, folders[0].Id); err != nil {
		ctx.WithFields(log.Fields{
			"folderId": folders[0].Id,
			"err":      err,
		}).Error("folderUsecase.RefreshCount failed")
		return err
	}
	return nil
}

func (u *erc721EventUseCase) buildTransferActivity(chainId domain.ChainId, event *contract.TransferEvent, lMeta *domain.LogMeta) *account.ActivityHistory {
	activity := &account.ActivityHistory{
		ChainId:         chainId,
		ContractAddress: lMeta.ContractAddress,
		Type:            account.ActivityHistoryTypeTransfer,
		TokenId:         event.TokenId,
		Account:         event.From,
		To:              event.To,
		Quantity:        "1",
		Price:           "0",
		BlockNumber:     lMeta.BlockNumber,
		TxHash:          lMeta.TxHash,
		LogIndex:        int64(lMeta.LogIndex),
		Time:            lMeta.BlockTime,
		Source:          account.SourceX,
	}

	if event.From == domain.EmptyAddress {
		activity.Type = account.ActivityHistoryTypeMint
	}
	return activity
}

// This use case is only for reindex
type erc721TransferActivityUseCase struct {
	*erc721EventUseCase
}

func NewErc721TransferActivityUseCase(cfg *Erc721EventUseCaseCfg) contract.Erc721EventUseCase {
	return &erc721TransferActivityUseCase{
		erc721EventUseCase: &erc721EventUseCase{
			nftitem:                   cfg.Nftitem,
			erc721:                    cfg.Erc721,
			token:                     cfg.Token,
			collectionRepo:            cfg.CollectionRepo,
			folderRepo:                cfg.FolderRepo,
			folderNftRelationshipRepo: cfg.FolderNftRelationshipRepo,
			activityHistoryRepo:       cfg.ActivityHistoryRepo,
			folderUsecase:             cfg.FolderUsecase,
		},
	}
}

func (u *erc721TransferActivityUseCase) Transfer(ctx bCtx.Ctx, chainId domain.ChainId, event *contract.TransferEvent, lMeta *domain.LogMeta) error {
	activity := u.buildTransferActivity(chainId, event, lMeta)
	return u.activityHistoryRepo.InsertTransferActivityIfNotExists(ctx, activity)
}
