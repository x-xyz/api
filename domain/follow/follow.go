package follow

import (
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
)

type Follow struct {
	From domain.Address `bson:"from"`
	To   domain.Address `bson:"to"`
}

type findAllOptions struct {
	SortBy  *string         `bson:"-"`
	SortDir *domain.SortDir `bson:"-"`
	Offset  *int32          `bson:"-"`
	Limit   *int32          `bson:"-"`
	From    *domain.Address `bson:"from"`
	To      *domain.Address `bson:"to"`
}

type FindAllOptions func(*findAllOptions) error

func GetFindAllOptions(opts ...FindAllOptions) (findAllOptions, error) {
	res := findAllOptions{}

	for _, opt := range opts {
		if err := opt(&res); err != nil {
			return res, err
		}
	}

	return res, nil
}

func WithSort(sortby string, sortdir domain.SortDir) FindAllOptions {
	return func(options *findAllOptions) error {
		options.SortBy = &sortby
		options.SortDir = &sortdir
		return nil
	}
}

func WithPagination(offset int32, limit int32) FindAllOptions {
	return func(options *findAllOptions) error {
		options.Offset = &offset
		options.Limit = &limit
		return nil
	}
}

func WithFrom(from domain.Address) FindAllOptions {
	return func(options *findAllOptions) error {
		from := from.ToLower()
		options.From = &from
		return nil
	}
}

func WithTo(to domain.Address) FindAllOptions {
	return func(options *findAllOptions) error {
		to := to.ToLower()
		options.To = &to
		return nil
	}
}

type Repo interface {
	Upsert(c ctx.Ctx, from, to domain.Address) error
	Remove(c ctx.Ctx, from, to domain.Address) error
	FindAll(c ctx.Ctx, opts ...FindAllOptions) ([]*Follow, error)
	Count(c ctx.Ctx, opts ...FindAllOptions) (int, error)
	// return nil if not found
	FindOne(c ctx.Ctx, from, to domain.Address) (*Follow, error)
}

type Usecase interface {
	Follow(c ctx.Ctx, address domain.Address, toAddress domain.Address) error
	Unfollow(c ctx.Ctx, address domain.Address, toAddress domain.Address) error
	GetFollowers(c ctx.Ctx, address domain.Address) ([]domain.Address, error)
	GetFollowerCount(c ctx.Ctx, address domain.Address) (int, error)
	GetFollowings(c ctx.Ctx, address domain.Address) ([]domain.Address, error)
	GetFollowingCount(c ctx.Ctx, address domain.Address) (int, error)
	IsFollowing(c ctx.Ctx, from, to domain.Address) (bool, error)
}
