package usecase

import (
	"math/big"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/base/ptr"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/account"
)

type orderNonceUCImpl struct {
	repo account.OrderNonceRepo
}

func NewOrderNonceUseCase(repo account.OrderNonceRepo) account.OrderNonceUseCase {
	return &orderNonceUCImpl{repo}
}

func (im *orderNonceUCImpl) FindOne(ctx ctx.Ctx, id account.OrderNonceId) (*account.OrderNonce, error) {
	return im.repo.FindOne(ctx, id)
}

func (im *orderNonceUCImpl) UseAvailableNonce(ctx ctx.Ctx, id account.OrderNonceId) (string, error) {
	orderNonce, err := im.getOrCreate(ctx, id)
	if err != nil {
		return "", err
	}
	availableNonce := orderNonce.NextAvailableNonce
	newNonce, err := incString(availableNonce)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":   err,
			"nonce": availableNonce,
		}).Error("incString failed")
		return "", err
	}
	patchable := account.OrderNoncePatchable{NextAvailableNonce: ptr.String(newNonce)}
	if err := im.repo.Update(ctx, id, patchable); err != nil {
		ctx.WithFields(log.Fields{
			"err":       err,
			"id":        id,
			"patchable": patchable,
		}).Error("repo.Update failed")
		return "", err
	}
	return availableNonce, nil
}

func (im *orderNonceUCImpl) UpdateMinValidOrderNonce(ctx ctx.Ctx, id account.OrderNonceId, nonce string) error {
	if _, err := im.getOrCreate(ctx, id); err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
			"id":  id,
		}).Error("im.getOrCreate failed")
		return err
	}

	if _, err := domain.ToBigInt([]string{nonce}); err != nil {
		ctx.WithFields(log.Fields{
			"err":   err,
			"nonce": nonce,
		}).Error("ToBigInt failed")
		return err
	}

	patchable := account.OrderNoncePatchable{MinValidOrderNonce: ptr.String(nonce)}
	if err := im.repo.Update(ctx, id, patchable); err != nil {
		ctx.WithFields(log.Fields{
			"err":       err,
			"id":        id,
			"patchable": patchable,
		}).Error("repo.Update failed")
		return err
	}
	return nil
}

func (im *orderNonceUCImpl) UpdateAvailableNonceIfNeeded(ctx ctx.Ctx, id account.OrderNonceId, usedNonce string) error {
	orderNonce, err := im.getOrCreate(ctx, id)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
			"id":  id,
		}).Error("im.getOrCreate failed")
		return err
	}

	numsStr := []string{orderNonce.NextAvailableNonce, usedNonce}
	nums, err := domain.ToBigInt(numsStr)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":     err,
			"numsStr": numsStr,
		}).Error("ToBigInt failed")
		return err
	}
	if nums[0].Cmp(nums[1]) > 0 { // nextAvailableNonce > usedNonce
		return nil
	}

	newNonce := new(big.Int).Add(nums[1], domain.Big1)
	patchable := account.OrderNoncePatchable{NextAvailableNonce: ptr.String(newNonce.String())}
	if err := im.repo.Update(ctx, id, patchable); err != nil {
		ctx.WithFields(log.Fields{
			"err":       err,
			"id":        id,
			"patchable": patchable,
		}).Error("repo.Update failed")
		return err
	}
	return nil
}

func (im *orderNonceUCImpl) getOrCreate(ctx ctx.Ctx, id account.OrderNonceId) (*account.OrderNonce, error) {
	orderNonce, err := im.repo.FindOne(ctx, id)
	if err == nil {
		return orderNonce, nil
	} else if err != domain.ErrNotFound {
		ctx.WithFields(log.Fields{
			"err": err,
			"id":  id,
		}).Error("repo.FindOne failed")
		return nil, err
	}

	orderNonce = &account.OrderNonce{
		Address:            id.Address,
		ChainId:            id.ChainId,
		NextAvailableNonce: "0",
		MinValidOrderNonce: "0",
	}
	if err := im.repo.Upsert(ctx, orderNonce); err != nil {
		ctx.WithFields(log.Fields{
			"err":        err,
			"orderNonce": orderNonce,
		}).Error("repo.Upsert failed")
		return nil, err
	}
	return orderNonce, nil
}

func incString(num string) (string, error) {
	n, ok := new(big.Int).SetString(num, 10)
	if !ok {
		return "", domain.ErrInvalidNumberFormat
	}
	return n.Add(n, domain.Big1).String(), nil
}
