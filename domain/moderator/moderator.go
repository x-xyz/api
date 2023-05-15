package moderator

import (
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
)

type Moderator struct {
	Name    string         `json:"name" bson:"name"`
	Address domain.Address `json:"address" bson:"address"`
}

type Repo interface {
	FindAll(c ctx.Ctx) ([]*Moderator, error)
	FindOne(c ctx.Ctx, address domain.Address) (*Moderator, error)
	Create(c ctx.Ctx, value Moderator) error
	Delete(c ctx.Ctx, address domain.Address) error
}

type Usecase interface {
	FindAll(c ctx.Ctx) ([]*Moderator, error)
	Add(c ctx.Ctx, address domain.Address, name string) error
	Remove(c ctx.Ctx, address domain.Address) error
	IsModerator(c ctx.Ctx, address domain.Address) (bool, error)
}
