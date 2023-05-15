package usecase

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/viney-shih/goroutines"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/base/ptr"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/account"
	"github.com/x-xyz/goapi/domain/collection"
	"github.com/x-xyz/goapi/domain/erc1155"
	"github.com/x-xyz/goapi/domain/keys"
	"github.com/x-xyz/goapi/domain/nftitem"
	"github.com/x-xyz/goapi/service/coingecko"
)

const day = 24 * time.Hour

type folderUsecaseImpl struct {
	folderRepo            account.FolderRepo
	relationRepo          account.FolderNftRelationshipRepo
	nftRepo               nftitem.Repo
	collectionRepo        collection.Repo
	floorPriceHistoryRepo collection.FloorPriceHistoryRepo
	coingecko             coingecko.Client
	erc1155HoldingRepo    erc1155.HoldingRepo

	workerPool *goroutines.Pool
}

func NewFolderUsecase(
	folderRepo account.FolderRepo,
	relationRepo account.FolderNftRelationshipRepo,
	nftRepo nftitem.Repo,
	collectionRepo collection.Repo,
	floorPriceHistoryRepo collection.FloorPriceHistoryRepo,
	coingecko coingecko.Client,
	erc1155HoldingRepo erc1155.HoldingRepo,
) account.FolderUseCase {
	return &folderUsecaseImpl{
		folderRepo:            folderRepo,
		relationRepo:          relationRepo,
		nftRepo:               nftRepo,
		collectionRepo:        collectionRepo,
		floorPriceHistoryRepo: floorPriceHistoryRepo,
		coingecko:             coingecko,
		erc1155HoldingRepo:    erc1155HoldingRepo,

		workerPool: goroutines.NewPool(32, goroutines.WithTaskQueueLength(1024), goroutines.WithPreAllocWorkers(8)),
	}
}

func (im *folderUsecaseImpl) Create(ctx ctx.Ctx, f *account.Folder) (string, error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		ctx.WithField("err", err).Error("failed to uuid.NewRandom")
		return "", err
	}

	id := uuid.String()
	f.Id = id
	if err := im.folderRepo.Insert(ctx, f); err != nil {
		ctx.WithFields(log.Fields{
			"folder": *f,
			"err":    err,
		}).Error("failed to Insert")
		return "", err
	}

	return id, nil
}

func (im *folderUsecaseImpl) Update(ctx ctx.Ctx, Id string, f *account.FolderUpdater, itemsInput []nftitem.Id) error {
	folder, err := im.folderRepo.Get(ctx, Id)
	if err != nil {
		ctx.WithFields(log.Fields{
			"folderId": Id,
			"err":      err,
		}).Error("failed to folderRepo.Get")
		return err
	}

	allNftsFromOwner, err := im.nftRepo.FindAll(ctx, nftitem.WithOwner(folder.Owner.ToLower()))
	if err != nil {
		ctx.WithFields(log.Fields{
			"owner": folder.Owner,
			"err":   err,
		}).Error("failed to nftRepo.FindAll")
		return err
	}

	allHoldingsFromOwner, err := im.erc1155HoldingRepo.FindAll(ctx, erc1155.WithOwner(folder.Owner.ToLower()))
	if err != nil {
		ctx.WithFields(log.Fields{
			"owner": folder.Owner,
			"err":   err,
		}).Error("failed to erc1155HoldingRepo.FindAll")
		return err
	}

	allNftMap := map[string]struct{}{}
	for _, nft := range allNftsFromOwner {
		allNftMap[nft.ToId().ToString()] = struct{}{}
	}
	for _, holding := range allHoldingsFromOwner {
		nftId := nftitem.Id{
			ChainId:         holding.ChainId,
			ContractAddress: holding.Address.ToLower(),
			TokenId:         holding.TokenId,
		}
		allNftMap[nftId.ToString()] = struct{}{}
	}

	items := []nftitem.Id{}
	for _, it := range itemsInput {
		if _, ok := allNftMap[it.ToString()]; ok {
			items = append(items, it)
		}
	}

	if err := im.relationRepo.DeleteAllRelationsByFolderID(ctx, Id); err != nil {
		ctx.WithFields(log.Fields{
			"folderId": Id,
			"err":      err,
		}).Error("failed to DeleteAllRelationsByFolderID")
		return err
	}

	if len(items) > 0 {
		if err := im.relationRepo.AddNftitemsToFolder(ctx, items, Id); err != nil {
			ctx.WithFields(log.Fields{
				"folderId": Id,
				"err":      err,
			}).Error("failed to AddNftitemsToFolder")
			return err
		}

		f.NftCount = ptr.Int(len(items))
		f.Cover = &items[0]
	} else {
		f.NftCount = ptr.Int(0)
		f.Cover = &nftitem.Id{}
	}

	if err := im.folderRepo.Update(ctx, Id, f); err != nil {
		ctx.WithFields(log.Fields{
			"folder": *f,
			"err":    err,
		}).Error("failed to update")
		return err
	}

	err = im.workerPool.ScheduleWithTimeout(3*time.Second, func() {
		if err := im.RefreshStat(ctx, Id); err != nil {
			ctx.WithFields(log.Fields{
				"folderId": Id,
				"err":      err,
			}).Error("failed to RefreshStat")
		}
	})
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":      err,
			"folderId": Id,
		}).Error("failed to ScheduleWithTimeout")
	}

	return nil
}

func (im *folderUsecaseImpl) GetFolder(ctx ctx.Ctx, folderId string) (*account.Folder, error) {
	folder, err := im.folderRepo.Get(ctx, folderId)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":      err,
			"folderId": folderId,
		}).Error("failed to Get folder")
		return nil, err
	}

	return folder, nil
}

func (im *folderUsecaseImpl) GetFolders(ctx ctx.Ctx, opts ...account.GetFoldersOptionsFunc) ([]*account.Folder, error) {
	folders, err := im.folderRepo.GetFolders(ctx, opts...)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to GetFolders")
		return nil, err
	}
	return folders, nil
}

func (im *folderUsecaseImpl) GetNFTsInFolder(ctx ctx.Ctx, folderId string) ([]*nftitem.NftitemWith1155Balance, error) {
	folder, err := im.folderRepo.Get(ctx, folderId)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":      err,
			"folderId": folderId,
		}).Error("failed to folderRepo.Get")
		return nil, err
	}

	relations, err := im.relationRepo.GetAllRelations(ctx, account.WithFolderId(folderId))
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":      err,
			"folderId": folderId,
		}).Error("failed to GetAllRelationsByFolderID")
		return nil, err
	}

	if len(relations) == 0 {
		return []*nftitem.NftitemWith1155Balance{}, nil
	}

	errCh := make(chan error, len(relations))
	nfts := make([]*nftitem.NftitemWith1155Balance, len(relations))
	for ind := range relations {
		go func(ind int) {
			relation := relations[ind]
			item, err := im.nftRepo.FindOne(ctx, relation.ChainId, relation.ContractAddress.ToLower(), relation.TokenId)
			if err == domain.ErrNotFound {
				// if nftitem not found delete relation
				ctx.WithFields(log.Fields{
					"folderId": folderId,
					"nftId":    *relation.ToNftItemId(),
				}).Info("nftitem not found")
				err := im.relationRepo.DeleteAll(ctx, account.WithNftitemId(*relation.ToNftItemId()))
				if err != nil {
					ctx.WithFields(log.Fields{
						"err": err,
					}).Error("failed to relationRepo.DeleteAll")
				}
				nfts[ind] = nil
				errCh <- nil
				return
			}
			if err != nil {
				ctx.WithFields(log.Fields{
					"err":      err,
					"folderId": folderId,
					"nftId":    *relation.ToNftItemId(),
				}).Error("failed to nftRepo.FindOne")
				errCh <- err
				return
			}

			nfts[ind] = &nftitem.NftitemWith1155Balance{NftItem: *item}

			if item.TokenType == 1155 {
				holdingId := erc1155.HoldingId{
					ChainId: relation.ChainId,
					Address: relation.ContractAddress.ToLower(),
					TokenId: relation.TokenId,
					Owner:   folder.Owner.ToLower(),
				}
				holding, err := im.erc1155HoldingRepo.FindOne(ctx, holdingId)
				if err != nil {
					ctx.WithFields(log.Fields{
						"err":       err,
						"folderId":  folderId,
						"holdingId": holdingId,
					}).Error("failed to erc1155HoldingRepo.FindOne")
					errCh <- err
					return
				}
				nfts[ind].Balance = int(holding.Balance)
			}

			errCh <- nil
		}(ind)
	}

	for i := 0; i < len(relations); i++ {
		err := <-errCh
		if err != nil {
			return nil, err
		}
	}

	// filter out nil nft
	res := []*nftitem.NftitemWith1155Balance{}
	for _, nft := range nfts {
		if nft != nil {
			res = append(res, nft)
		}
	}

	return res, nil
}

func (im *folderUsecaseImpl) RefreshStat(ctx ctx.Ctx, folderId string) error {
	folder, err := im.folderRepo.Get(ctx, folderId)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":      err,
			"folderId": folderId,
		}).Error("failed to Get folder")
		return err
	}

	nfts, err := im.GetNFTsInFolder(ctx, folderId)
	if err != nil {
		ctx.WithFields(log.Fields{
			"folderId": folderId,
			"err":      err,
		}).Error("GetNFTsInFolder failed")
		return err
	}

	ethPrice, err := im.coingecko.GetPrice(ctx, "ethereum")
	if err != nil {
		ctx.WithFields(log.Fields{
			"folderId": folderId,
			"err":      err,
		}).Error("failed to coingecko.GetPrice")
		return err
	}

	floorPriceInUsd := float64(0)
	totalValueInUsd := float64(0)
	previousTotalValueInUsd := float64(0)
	instantLiquidityInUsd := float64(0)
	floors := make(map[string]float64)
	previousFloors := make(map[string]float64)
	yesterday := time.Now().Add(-day).Truncate(day)
	nftcount := len(nfts)
	for _, nft := range nfts {
		balance := 1
		if nft.TokenType == 1155 {
			holdingId := erc1155.HoldingId{
				ChainId: nft.ChainId,
				Address: nft.ContractAddress.ToLower(),
				TokenId: nft.TokenId,
				Owner:   folder.Owner.ToLower(),
			}
			holding, err := im.erc1155HoldingRepo.FindOne(ctx, holdingId)
			if err == domain.ErrNotFound {
				ctx.WithFields(log.Fields{
					"folderId":  folderId,
					"owner":     folder.Owner.ToLower(),
					"holdingId": holdingId,
				}).Info("erc1155HoldingRepo.FindOne not found")

				// delete relations if not found
				err := im.DeleteRelationFromAllFolders(ctx, folder.Owner.ToLower(), *nft.ToId())
				if err != nil {
					ctx.WithFields(log.Fields{
						"err": err,
					}).Error("failed to DeleteRelationFromAllFolders")
				}
				nftcount -= 1
				continue
			}
			if err != nil {
				ctx.WithFields(log.Fields{
					"err":       err,
					"holdingId": holdingId,
				}).Error("failed to erc1155HoldingRepo.FindOne")
				return err
			}
			balance = int(holding.Balance)
		}

		fKey := keys.RedisKey(fmt.Sprintf("%d", nft.ChainId), nft.ContractAddress.ToLowerStr())
		collectionFloorInUsd, ok := floors[fKey]
		if !ok {
			colId := collection.CollectionId{ChainId: nft.ChainId, Address: nft.ContractAddress}
			col, err := im.collectionRepo.FindOne(ctx, colId)
			if err == domain.ErrNotFound {
				// it is possible that we indexed the contract but not added the collection, this happended mostly at the beginning of x
				continue
			} else if err != nil {
				ctx.WithFields(log.Fields{
					"coldId": colId,
					"err":    err,
				}).Error("collectionRepo.FindOne failed")
				return err
			}
			collectionFloorInUsd = col.OpenseaFloorPriceInNative * ethPrice.InexactFloat64()
			floors[fKey] = collectionFloorInUsd
		}

		previousCollectionFloorInUsd, ok := previousFloors[fKey]
		if !ok {
			fpId := collection.FloorPriceId{ChainId: nft.ChainId, Address: nft.ContractAddress, Date: yesterday}
			fp, err := im.floorPriceHistoryRepo.FindOne(ctx, fpId)
			if err == nil {
				previousCollectionFloorInUsd = fp.OpenseaPriceInUsd
			} else if err == domain.ErrNotFound {
				previousCollectionFloorInUsd = 0
			} else {
				ctx.WithFields(log.Fields{
					"fpId": fpId,
					"err":  err,
				}).Error("floorPriceHistoryRepo.FindOne failed")
				return err
			}
			previousFloors[fKey] = previousCollectionFloorInUsd
		}

		if collectionFloorInUsd > 0 && (floorPriceInUsd == 0 || collectionFloorInUsd < floorPriceInUsd) {
			floorPriceInUsd = collectionFloorInUsd
		}

		totalValueInUsd += collectionFloorInUsd * float64(balance)
		previousTotalValueInUsd += previousCollectionFloorInUsd * float64(balance)
		instantLiquidityInUsd += nft.InstantLiquidityInUsd
	}

	totalValueMovement := float64(0)
	if previousTotalValueInUsd != 0 {
		totalValueMovement = (totalValueInUsd - previousTotalValueInUsd) / previousTotalValueInUsd
	}

	updater := &account.FolderUpdater{
		FloorPriceInUsd:       &floorPriceInUsd,
		TotalValueInUsd:       &totalValueInUsd,
		InstantLiquidityInUsd: &instantLiquidityInUsd,
		TotalValueMovement:    &totalValueMovement,
		NftCount:              ptr.Int(nftcount),
		CollectionCount:       ptr.Int(len(floors)),
	}
	if err := im.folderRepo.Update(ctx, folderId, updater); err != nil {
		ctx.WithFields(log.Fields{
			"folderId": folderId,
			"updater":  updater,
			"err":      err,
		})
		return err
	}

	return nil
}

func (im *folderUsecaseImpl) RefreshCount(ctx ctx.Ctx, folderId string) error {
	_, err := im.folderRepo.Get(ctx, folderId)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":      err,
			"folderId": folderId,
		}).Error("failed to Get folder")
		return err
	}

	count, err := im.relationRepo.Count(ctx, account.WithFolderId(folderId))
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":      err,
			"folderId": folderId,
		}).Error("failed to relationRepo.Count")
		return err
	}

	updater := account.FolderUpdater{
		NftCount: ptr.Int(count),
	}

	if err := im.folderRepo.Update(ctx, folderId, &updater); err != nil {
		ctx.WithFields(log.Fields{
			"err":      err,
			"folderId": folderId,
			"updater":  updater,
		}).Error("failed to folderRepo.Update")
		return err
	}

	return nil
}

func (im *folderUsecaseImpl) Delete(c ctx.Ctx, folderId string) error {
	if err := im.folderRepo.Delete(c, folderId); err != nil {
		c.WithFields(log.Fields{
			"folderId": folderId,
			"err":      err,
		}).Error("failed to folderRepo.Delete")
		return err
	}

	return nil
}

func (im *folderUsecaseImpl) MarkNftPrivate(c ctx.Ctx, owner domain.Address, marks []nftitem.Id, unmarks []nftitem.Id) error {
	builtInFolders, err := im.folderRepo.GetFolders(c, account.WithOwner(owner), account.WithBuiltIn(true))
	if err != nil {
		c.WithFields(log.Fields{
			"err":   err,
			"owner": owner,
		}).Error("failed to GetFolders")
		return err
	}

	publicFolder := account.Folder{}
	privateFolder := account.Folder{}
	for _, f := range builtInFolders {
		if f.IsPrivate {
			privateFolder = *f
		} else {
			publicFolder = *f
		}
	}

	// check nfts' ownership
	allNfts, err := im.nftRepo.FindAll(c, nftitem.WithOwner(owner))
	if err != nil {
		c.WithFields(log.Fields{
			"err":   err,
			"owner": owner,
		}).Error("failed to nftRepo.FindAll")
		return err
	}

	allHoldingsFromOwner, err := im.erc1155HoldingRepo.FindAll(c, erc1155.WithOwner(owner.ToLower()))
	if err != nil {
		c.WithFields(log.Fields{
			"owner": owner,
			"err":   err,
		}).Error("failed to erc1155HoldingRepo.FindAll")
		return err
	}

	allNftsMap := map[string]struct{}{}
	for _, nft := range allNfts {
		allNftsMap[nft.ToId().ToString()] = struct{}{}
	}
	for _, holding := range allHoldingsFromOwner {
		nftId := nftitem.Id{
			ChainId:         holding.ChainId,
			ContractAddress: holding.Address.ToLower(),
			TokenId:         holding.TokenId,
		}
		allNftsMap[nftId.ToString()] = struct{}{}
	}

	if len(allNftsMap) == 0 {
		return nil
	}

	marksFiltered := []nftitem.Id{}
	for _, id := range marks {
		if _, ok := allNftsMap[id.ToString()]; ok {
			marksFiltered = append(marksFiltered, id)
		}
	}

	unmarksFiltered := []nftitem.Id{}
	for _, id := range unmarks {
		if _, ok := allNftsMap[id.ToString()]; ok {
			unmarksFiltered = append(unmarksFiltered, id)
		}
	}

	err = im.relationRepo.MoveNftitems(c, marksFiltered, publicFolder.Id, privateFolder.Id)
	if err != nil {
		c.WithFields(log.Fields{
			"err": err,
		}).Error("failed to MoveNftitems")
		return err
	}

	err = im.relationRepo.MoveNftitems(c, unmarksFiltered, privateFolder.Id, publicFolder.Id)
	if err != nil {
		c.WithFields(log.Fields{
			"err": err,
		}).Error("failed to MoveNftitems")
		return err
	}

	publicCount, err := im.relationRepo.Count(c, account.WithFolderId(publicFolder.Id))
	if err != nil {
		c.WithFields(log.Fields{
			"err": err,
		}).Error("failed to relationRepo.Count")
		return err
	}

	err = im.folderRepo.Update(c, publicFolder.Id, &account.FolderUpdater{
		NftCount: ptr.Int(publicCount),
	})
	if err != nil {
		c.WithFields(log.Fields{
			"err": err,
		}).Error("failed to folderRepo.Update")
		return err
	}

	privateCount, err := im.relationRepo.Count(c, account.WithFolderId(privateFolder.Id))
	if err != nil {
		c.WithFields(log.Fields{
			"err": err,
		}).Error("failed to relationRepo.Count")
		return err
	}

	err = im.folderRepo.Update(c, privateFolder.Id, &account.FolderUpdater{
		NftCount: ptr.Int(privateCount),
	})
	if err != nil {
		c.WithFields(log.Fields{
			"err": err,
		}).Error("failed to folderRepo.Update")
		return err
	}

	err = im.workerPool.ScheduleWithTimeout(3*time.Second, func() {
		if err := im.RefreshStat(c, publicFolder.Id); err != nil {
			c.WithFields(log.Fields{
				"folderId": publicFolder.Id,
				"err":      err,
			}).Error("failed to RefreshStat")
		}

		if err := im.RefreshStat(c, privateFolder.Id); err != nil {
			c.WithFields(log.Fields{
				"folderId": privateFolder.Id,
				"err":      err,
			}).Error("failed to RefreshStat")
		}
	})
	if err != nil {
		c.WithFields(log.Fields{
			"err": err,
		}).Error("failed to ScheduleWithTimeout")
	}

	return nil
}

func (im *folderUsecaseImpl) InitBuiltInFolders(ctx ctx.Ctx, owner domain.Address) error {
	folders, err := im.GetFolders(ctx, account.WithOwner(owner), account.WithBuiltIn(true))
	if err != nil {
		ctx.WithFields(log.Fields{
			"owner": owner,
			"err":   err,
		}).Error("failed to GetFolders")
		return err
	}

	var publicFolder *account.Folder
	var privateFolder *account.Folder

	for _, f := range folders {
		if f.IsBuiltIn && !f.IsPrivate {
			publicFolder = f
		}
		if f.IsBuiltIn && f.IsPrivate {
			privateFolder = f
		}
	}

	if publicFolder == nil {
		id, _ := uuid.NewRandom()
		publicFolder = &account.Folder{
			Id:                    id.String(),
			Name:                  "public",
			IsPrivate:             false,
			IsBuiltIn:             true,
			FloorPriceInUsd:       0,
			TotalValueInUsd:       0,
			InstantLiquidityInUsd: 0,
			Owner:                 owner,
			CreatedAt:             time.Now(),
		}
		err := im.folderRepo.Insert(ctx, publicFolder)
		if err != nil {
			ctx.WithFields(log.Fields{
				"err":    err,
				"folder": *publicFolder,
			}).Error("failed to folderRepo.Insert")
			return err
		}
	}

	if privateFolder == nil {
		id, _ := uuid.NewRandom()
		privateFolder = &account.Folder{
			Id:                    id.String(),
			Name:                  "private",
			IsPrivate:             true,
			IsBuiltIn:             true,
			FloorPriceInUsd:       0,
			TotalValueInUsd:       0,
			InstantLiquidityInUsd: 0,
			Owner:                 owner,
			CreatedAt:             time.Now(),
		}
		err := im.folderRepo.Insert(ctx, privateFolder)
		if err != nil {
			ctx.WithFields(log.Fields{
				"err":    err,
				"folder": *privateFolder,
			}).Error("failed to folderRepo.Insert")
			return err
		}
	}

	findOpts := []nftitem.FindAllOptionsFunc{nftitem.WithOwner(owner)}

	holdings, err := im.erc1155HoldingRepo.FindAll(ctx, erc1155.WithOwner(owner))
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":   err,
			"owner": owner,
		}).Error("failed to FindAll")
		return err
	}
	holdingNftitemIds := []nftitem.Id{}
	for _, holding := range holdings {
		if holding.Balance > 0 {
			nftid := nftitem.Id{
				ChainId:         holding.ChainId,
				ContractAddress: holding.Address.ToLower(),
				TokenId:         holding.TokenId,
			}
			holdingNftitemIds = append(holdingNftitemIds, nftid)
		}
	}
	if len(holdingNftitemIds) > 0 {
		findOpts = append(findOpts, nftitem.WithHoldingIds(holdingNftitemIds))
	}

	allNfts, err := im.nftRepo.FindAll(ctx, findOpts...)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":   err,
			"owner": owner,
		}).Error("failed to nftRepo.FindAll")
		return err
	}

	nftCount := len(allNfts)

	if nftCount > 0 {
		itemIds := []nftitem.Id{}
		for _, nft := range allNfts {
			itemIds = append(itemIds, *nft.ToId())
		}
		err := im.relationRepo.MoveNftitems(ctx, itemIds, privateFolder.Id, publicFolder.Id)
		if err != nil {
			ctx.WithFields(log.Fields{
				"err": err,
			}).Error("failed to relationRepo.MoveNftitems")
			return err
		}
	}

	err = im.folderRepo.Update(ctx, publicFolder.Id, &account.FolderUpdater{
		NftCount: &nftCount,
	})
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to Update")
		return err
	}

	err = im.workerPool.ScheduleWithTimeout(3*time.Second, func() {
		if err := im.RefreshStat(ctx, publicFolder.Id); err != nil {
			ctx.WithFields(log.Fields{
				"folderId": publicFolder.Id,
				"err":      err,
			}).Error("failed to RefreshStat")
		}
	})
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":      err,
			"folderId": publicFolder.Id,
		}).Error("failed to ScheduleWithTimeout")
	}

	return nil
}

func (im *folderUsecaseImpl) DeleteRelationFromAllFolders(ctx ctx.Ctx, owner domain.Address, nftitemId nftitem.Id) error {
	folders, err := im.folderRepo.GetFolders(ctx, account.WithOwner(owner.ToLower()))
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to folderRepo.GetFolders")
		return err
	}

	if len(folders) == 0 {
		// folders not found, user not registered
		// return without error
		return nil
	}

	folderIds := []string{}
	for _, folder := range folders {
		folderIds = append(folderIds, folder.Id)
	}

	err = im.relationRepo.DeleteAll(ctx, account.WithFolderIds(folderIds), account.WithNftitemId(nftitemId))
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to relationRepo.DeleteAll")
		return err
	}

	return nil
}

func (im *folderUsecaseImpl) AddNftToPublicFolder(ctx ctx.Ctx, owner domain.Address, nftitemId nftitem.Id) error {
	folders, err := im.folderRepo.GetFolders(ctx, account.WithBuiltIn(true), account.WithOwner(owner.ToLower()))
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":   err,
			"owner": owner,
		}).Error("failed to folderRepo.GetFolders")
		return err
	}

	if len(folders) == 0 {
		// public folder not found, user not registered
		// return without error
		return nil
	}

	var publicFolder *account.Folder
	var privateFolder *account.Folder
	for _, folder := range folders {
		if folder.IsPrivate {
			privateFolder = folder
		} else {
			publicFolder = folder
		}
	}

	err = im.relationRepo.MoveNftitems(ctx, []nftitem.Id{nftitemId}, privateFolder.Id, publicFolder.Id)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to relationRepo.AddNftitemsToFolder")
		return err
	}

	return nil
}
