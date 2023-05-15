package usecase

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"time"

	"github.com/viney-shih/goroutines"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/ethereum"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/base/ptr"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/account"
	"github.com/x-xyz/goapi/domain/collection"
	"github.com/x-xyz/goapi/domain/file"
	"github.com/x-xyz/goapi/domain/follow"
	"github.com/x-xyz/goapi/domain/like"
	"github.com/x-xyz/goapi/domain/moderator"
	"github.com/x-xyz/goapi/domain/nftitem"
	"github.com/x-xyz/goapi/domain/token"
	"github.com/x-xyz/goapi/service/pinata"
)

const (
	nonceRange   = int32(9999999)
	invalidNonce = int32(-1)
)

type AccountUseCaseCfg struct {
	Repo                    account.Repo
	NotificationSettingRepo account.NotificationSettingsRepo
	FollowUC                follow.Usecase
	ModeratorUC             moderator.Usecase
	FileUC                  file.Usecase
	NftitemRepo             nftitem.Repo
	TokenUC                 token.Usecase
	LikeUC                  like.Usecase
	SignatureMsg            string
	CollectionUC            collection.Usecase
	ActivityRepo            account.ActivityHistoryRepo
	FolderUC                account.FolderUseCase
}

type impl struct {
	repo         account.Repo
	nsRepo       account.NotificationSettingsRepo
	nftitem      nftitem.Repo
	activityRepo account.ActivityHistoryRepo
	follow       follow.Usecase
	moderator    moderator.Usecase
	file         file.Usecase
	token        token.Usecase
	like         like.Usecase
	signatureMsg string
	collection   collection.Usecase
	folder       account.FolderUseCase
}

// New creates account usecase
func New(cfg *AccountUseCaseCfg) account.Usecase {
	return &impl{
		repo:         cfg.Repo,
		nsRepo:       cfg.NotificationSettingRepo,
		follow:       cfg.FollowUC,
		file:         cfg.FileUC,
		moderator:    cfg.ModeratorUC,
		token:        cfg.TokenUC,
		like:         cfg.LikeUC,
		nftitem:      cfg.NftitemRepo,
		signatureMsg: cfg.SignatureMsg,
		collection:   cfg.CollectionUC,
		activityRepo: cfg.ActivityRepo,
		folder:       cfg.FolderUC,
	}
}

func (im *impl) Get(c ctx.Ctx, address domain.Address) (*account.Info, error) {
	a, err := im.repo.Get(c, address)
	if err != nil {
		c.WithFields(log.Fields{
			"address": address,
			"err":     err,
		}).Error("get address error")
		return nil, err
	}
	return im.accountToInfo(c, a)
}

func (im *impl) Update(c ctx.Ctx, address domain.Address, a *account.Updater) (*account.Info, error) {
	c = ctx.WithValues(c, map[string]interface{}{
		"address":    address,
		"alias":      a.Alias,
		"email":      a.Email,
		"bio":        a.Bio,
		"imageHash":  a.ImageHash,
		"bannerHash": a.BannerHash,
	})
	a.UpdatedAt = time.Now()
	if err := im.repo.Update(c, address, a); err != nil {
		c.WithField("err", err).Error("repo.Update failed")
		return nil, err
	}
	return im.Get(c, address)
}

func (im *impl) Create(c ctx.Ctx, address domain.Address) (*account.Info, error) {
	c = ctx.WithValues(c, map[string]interface{}{
		"address": address,
	})
	new, err := im.create(c, address)
	if err != nil {
		return nil, err
	}
	return im.accountToInfo(c, new)
}

func (im *impl) create(c ctx.Ctx, address domain.Address) (*account.Account, error) {
	now := time.Now()
	new := &account.Account{
		Address:       address,
		Nonce:         invalidNonce,
		IsAppropriate: true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := im.repo.Insert(c, new); err != nil {
		c.WithField("err", err).Error("repo.Insert failed")
		return nil, err
	}
	return new, nil
}

func (im *impl) getOrCreate(c ctx.Ctx, address domain.Address) (*account.Account, error) {
	_account, err := im.repo.Get(c, address)
	if err == domain.ErrNotFound {
		_account, err = im.create(c, address)
		if err != nil {
			c.WithFields(log.Fields{
				"err": err,
			}).Error("im.create failed")
			return nil, err
		}
		return _account, nil
	} else if err != nil {
		c.WithFields(log.Fields{
			"err": err,
		}).Error("repo.Get failed")
		return nil, err
	}
	return _account, nil
}

func (im *impl) UpdateBanner(c ctx.Ctx, address domain.Address, imgData string) (string, error) {
	hash, err := im.file.Upload(c, imgData, pinata.PinOptions{
		Metadata: &pinata.PinataMetadata{
			Name:      string(address),
			KeyValues: map[string]interface{}{},
		},
		Options: &pinata.PinataOptions{
			CidVersion: pinata.CidVersion_0,
		},
	})
	if err != nil {
		c.WithField("err", err).Error("file.Upload failed")
		return "", err
	}

	if err := im.repo.Update(c, address, &account.Updater{
		BannerHash: &hash,
	}); err != nil {
		c.WithField("err", err).Error("repo.Update failed")
		return "", err
	}
	return hash, nil
}

func (im *impl) UpdateAvatar(c ctx.Ctx, address domain.Address, imgData string) (string, error) {
	a, err := im.repo.Get(c, address)
	if err != nil {
		c.WithFields(log.Fields{
			"address": address,
			"err":     err,
		}).Error("get address error")
		return "", err
	}

	hash, err := im.file.Upload(c, imgData, pinata.PinOptions{
		Metadata: &pinata.PinataMetadata{
			Name: a.Alias + string(address) + "avatar",
			KeyValues: map[string]interface{}{
				"address":  a.Address,
				"userName": a.Alias,
			},
		},
		Options: &pinata.PinataOptions{
			CidVersion: pinata.CidVersion_0,
		},
	})
	if err != nil {
		c.WithField("err", err).Error("file.Upload failed")
		return "", err
	}

	if err := im.repo.Update(c, address, &account.Updater{
		ImageHash: &hash,
	}); err != nil {
		c.WithField("err", err).Error("repo.Update failed")
		return "", err
	}
	return hash, nil
}

// TODO: using Redis to store nonce instead
func (im *impl) GenerateNonce(c ctx.Ctx, address domain.Address) (int32, error) {
	c = ctx.WithValue(c, "address", address)
	if _, err := im.Get(c, address); err != nil && err != domain.ErrNotFound {
		c.WithField("err", err).Error("get account failed")
		return 0, err
	} else if err == domain.ErrNotFound {
		// if the account doesn't exist, create an empty account
		if _, err := im.Create(c, address); err != nil {
			c.WithField("err", err).Error("im.Create account failed")
			return 0, err
		}
		c.Info("created new account")
	}

	nonce := im.genNonce()
	if err := im.repo.Update(c, address, &account.Updater{
		Nonce: nonce,
	}); err != nil {
		c.WithField("err", err).Error("repo.Update failed")
		return 0, err
	}
	return nonce, nil
}

func (im *impl) makeMessageWithNonce(nonce string) []byte {
	return []byte(fmt.Sprintf(im.signatureMsg, nonce))
}

// TODO: cloud be migrated to Auth middleware after we implemented nonce using Redis
func (im *impl) ValidateSignature(c ctx.Ctx, address domain.Address, signature string) error {
	c = ctx.WithValues(c, map[string]interface{}{
		"address":   address,
		"signature": signature,
	})

	// get nonce and check is it valid
	a, err := im.repo.Get(c, address)
	if err != nil {
		c.WithField("err", err).Error("get address failed")
		return err
	}
	if a.Nonce == invalidNonce {
		return account.ErrInvalidNonce
	}

	// reset nonce after validated the signature
	defer im.repo.Update(c, address, &account.Updater{
		Nonce: invalidNonce,
	})

	// validate nonce
	msg := im.makeMessageWithNonce(strconv.Itoa(int(a.Nonce)))
	if isValid, err := ethereum.ValidateMsgSignature(msg, signature, string(address)); err != nil {
		c.WithField("err", err).Error("ValidateMsgSignature failed")
		return err
	} else if !isValid {
		return account.ErrInvalidSignature
	}
	return nil
}

func (im *impl) GetNotificationSettings(c ctx.Ctx, address domain.Address) (*account.NotificationSettings, error) {
	return im.nsRepo.Get(c, address)
}

func (im *impl) UpsertNotificationSettings(c ctx.Ctx, settings *account.NotificationSettings) (*account.NotificationSettings, error) {
	return im.nsRepo.Upsert(c, settings)
}

func (im *impl) Follow(c ctx.Ctx, address, toAddress domain.Address) error {
	// check followee exists
	if _, err := im.Get(c, toAddress); err != nil {
		c.WithField("err", err).Error("im.Get failed")
		return err
	}

	if err := im.follow.Follow(c, address, toAddress); err != nil {
		c.WithField("err", err).Error("follow.Follow failed")
		return err
	}
	return nil
}

func (im *impl) Unfollow(c ctx.Ctx, address, toAddress domain.Address) error {
	// check follower exists
	if _, err := im.Get(c, toAddress); err != nil {
		c.WithField("err", err).Error("im.Get failed")
		return err
	}
	if err := im.follow.Unfollow(c, address, toAddress); err != nil {
		c.WithField("err", err).Error("follow.Unfollow failed")
		return err
	}
	return nil
}

func (im *impl) getAccounts(c ctx.Ctx, addresses []domain.Address) ([]*account.Info, error) {
	// TODO: using cache to reduce some repo calls
	// TODO: handling some missing accounts
	accounts, err := im.repo.GetAccounts(c, addresses)
	if err != nil {
		c.WithField("err", err).Error("repo.GetAccounts failed")
		return nil, err
	}

	// batch get account infos
	b := goroutines.NewBatch(10, goroutines.WithBatchSize(len(accounts)))
	defer b.Close()
	for i := 0; i < len(accounts); i++ {
		idx := i
		b.Queue(func() (interface{}, error) {
			info, err := im.accountToInfo(c, accounts[idx])
			if err != nil {
				info = accounts[idx].ToInfo()
			}
			return info, nil
		})
	}
	b.QueueComplete()

	idx := 0
	infos := make([]*account.Info, len(accounts))
	for ret := range b.Results() {
		if ret.Error() != nil {
			c.WithField("err", err).Error("get account info error result")
			continue
		}
		infos[idx] = ret.Value().(*account.Info)
		idx++
	}
	return infos, nil
}

func (im *impl) IsFollowing(c ctx.Ctx, address, toAddress domain.Address) (bool, error) {
	isFollowing, err := im.follow.IsFollowing(c, address, toAddress)
	if err != nil {
		c.WithField("err", err).Error("follow.IsFollowing failed")
		return false, err
	}
	return isFollowing, err
}

func (im *impl) GetFollowers(c ctx.Ctx, address domain.Address) ([]*account.Info, error) {
	addresses, err := im.follow.GetFollowers(c, address)
	if err != nil {
		c.WithField("err", err).Error("follow.Followers failed")
		return nil, err
	}
	infos, err := im.getAccounts(c, addresses)
	if err != nil {
		c.WithField("err", err).Error("im.getAccounts failed")
		return nil, err
	}
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Address < infos[j].Address
	})
	return infos, nil
}

func (im *impl) GetFollowings(c ctx.Ctx, address domain.Address) ([]*account.Info, error) {
	addresses, err := im.follow.GetFollowings(c, address)
	if err != nil {
		c.WithField("err", err).Error("follow.GetFollowings failed")
		return nil, err
	}
	infos, err := im.getAccounts(c, addresses)
	if err != nil {
		c.WithField("err", err).Error("im.getAccounts failed")
		return nil, err
	}
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Address < infos[j].Address
	})
	return infos, nil
}

func (im *impl) Ban(c ctx.Ctx, address domain.Address) error {
	if err := im.repo.Update(c, address, &account.Updater{IsAppropriate: ptr.Bool(false)}); err != nil {
		c.WithField("err", err).WithField("address", address).Error("repo.Update failed")
		return err
	}
	return nil
}

func (im *impl) Unban(c ctx.Ctx, address domain.Address) error {
	if err := im.repo.Update(c, address, &account.Updater{IsAppropriate: ptr.Bool(true)}); err != nil {
		c.WithField("err", err).WithField("address", address).Error("repo.Update failed")
		return err
	}
	return nil
}

func (im *impl) GetActivities(c ctx.Ctx, address domain.Address, optFns ...account.FindActivityHistoryOptions) (*account.ActivityResult, error) {
	activityOpts := append(
		[]account.FindActivityHistoryOptions{
			account.ActivityHistoryWithTypes(
				account.ActivityHistoryTypeCreateOffer,
				account.ActivityHistoryTypeCancelOffer,
				account.ActivityHistoryTypeList,
				account.ActivityHistoryTypeCancelListing,
				account.ActivityHistoryTypePlaceBid,
				account.ActivityHistoryTypeBuy,
				account.ActivityHistoryTypeSold,
				account.ActivityHistoryTypeTransfer,
				account.ActivityHistoryTypeMint,
				account.ActivityHistoryTypeSale,
			),
		},
		optFns...,
	)

	activityOpts = append(activityOpts, account.ActivityHistoryWithAccount(address))

	res := &account.ActivityResult{}

	activities, err := im.activityRepo.FindActivities(c, activityOpts...)
	if err != nil {
		c.WithField("err", err).Error("activityRepo.FindActivities")
		return nil, err
	}

	count, err := im.activityRepo.CountActivities(c, activityOpts...)
	if err != nil {
		c.WithField("err", err).Error("activityRepo.CountActivities")
		return nil, err
	}

	for _, act := range activities {
		token, err := im.getSimpleToken(c, act.ChainId, act.ContractAddress, act.TokenId)
		if err != nil {
			c.WithField("err", err).Warn("getSimpleTokenAndOwner failed")
			continue
		}

		a, err := act.ToActivity()
		if err == account.ErrNotFoundActivityType {
			continue
		} else if err != nil {
			c.WithField("err", err).WithField("activityHistory", act).Warn("ToActivity failed")
			continue
		}

		a.Token = *token

		a.Owner = im.getSimpleAccount(c, act.Account)
		if a.Type == account.ActivityTypeTransfer || a.Type == account.ActivityTypeMint || a.Type == account.ActivityTypeSale2 {
			a.To = im.getSimpleAccount(c, act.To)
		}
		res.Activities = append(res.Activities, a)
	}

	res.Count = count

	return res, nil
}

func (im *impl) GetAccountStat(c ctx.Ctx, address domain.Address) (*account.AccountStat, error) {
	res := &account.AccountStat{}

	if count, err := im.like.GetLikedCount(c, address); err != nil {
		c.WithField("err", err).Error("like.GetLikedCount failed")
		return nil, err
	} else {
		res.Favorite = int32(count)
	}

	if items, err := im.nftitem.FindAll(c, nftitem.WithOwner(address)); err != nil {
		c.WithField("err", err).Error("nftitem.FindAll failed")
		return nil, err
	} else {
		countOfNfts := int32(0)
		countOfCollections := int32(0)
		countedCollection := map[string]bool{}

		for _, item := range items {
			countOfNfts += 1
			collection := strconv.Itoa(int(item.ChainId)) + ":" + string(item.ContractAddress)
			if _, ok := countedCollection[collection]; !ok {
				countedCollection[collection] = true
				countOfCollections += 1
			}
		}

		res.Single = countOfNfts
		res.Collections = countOfCollections
	}

	if cols, err := im.collection.FindAll(c, collection.WithOwner(address)); err != nil {
		c.WithField("err", err).Error("collection.FindAll failed")
		return nil, err
	} else {
		countOfNfts := int32(0)

		for _, col := range cols.Items {
			if cnt, err := im.nftitem.Count(c, nftitem.WithChainId(col.ChainId), nftitem.WithContractAddresses([]domain.Address{col.Erc721Address})); err != nil {
				c.WithFields(log.Fields{"err": err, "collection": col}).Error("nftitem.Count failed")
			} else {
				countOfNfts += int32(cnt)
			}
		}

		res.CreatedNfts = countOfNfts
		res.CreatedCollections = int32(cols.Count)
	}

	return res, nil
}

func (im *impl) GetAccountCollectionHoldings(c ctx.Ctx, address domain.Address) (*account.AccountCollectionHoldings, error) {
	res := &account.AccountCollectionHoldings{
		Collections:               make(map[account.CollectionId]int32),
		CollectionsHoldingBalance: make(map[account.CollectionId]int32),
	}

	items, err := im.token.SearchV2(c, token.WithBelongsTo(address.ToLower()))
	if err != nil {
		c.WithFields(log.Fields{
			"owner": address,
			"err":   err,
		}).Error("nftitem.FindAll failed")
		return nil, err
	}

	for _, item := range items.Items {
		id := account.CollectionId{ChainId: item.ChainId, Address: item.ContractAddress}
		res.Collections[id] += 1
		if item.TokenType == 1155 && item.Balance != nil {
			res.CollectionsHoldingBalance[id] += int32(*item.Balance)
		} else {
			res.CollectionsHoldingBalance[id] += 1
		}
	}

	return res, nil
}

func (im *impl) getSimpleToken(c ctx.Ctx, chainId domain.ChainId, contract domain.Address, tokenId domain.TokenId) (*nftitem.SimpleNftItem, error) {
	id := nftitem.Id{ChainId: chainId, ContractAddress: contract, TokenId: tokenId}

	token, err := im.token.FindOne(c, id)

	if err != nil {
		c.WithField("err", err).WithField("id", id).Warn("token.FindOne failed")
		return nil, err
	}

	return token.ToSimpleNftItem(), nil
}

func (im *impl) getSimpleAccount(c ctx.Ctx, address domain.Address) account.SimpleAccount {
	a, err := im.repo.Get(c, address)
	if err != nil {
		c.WithField("err", err).WithField("address", address).Warn("Get failed")
		return account.SimpleAccount{
			Address: address,
		}
	}
	sa := a.ToSimpleAccount()
	return *sa
}

func (im *impl) genNonce() int32 {
	return rand.Int31n(nonceRange)
}

func (im *impl) accountToInfo(c ctx.Ctx, a *account.Account) (*account.Info, error) {
	info := a.ToInfo()

	if count, err := im.follow.GetFollowerCount(c, info.Address); err != nil {
		c.WithField("err", err).Error("follow.GetFollowerCount failed")
		return nil, err
	} else {
		info.Followers = int32(count)
	}

	if count, err := im.follow.GetFollowingCount(c, info.Address); err != nil {
		c.WithField("err", err).Error("follow.GetFollowingCount failed")
		return nil, err
	} else {
		info.Followings = int32(count)
	}

	if isModerator, err := im.moderator.IsModerator(c, info.Address); err != nil {
		c.WithField("err", err).Error("moderator.IsModerator failed")
		return nil, err
	} else {
		info.IsModerator = isModerator
	}

	return info.Sanitized(), nil
}
