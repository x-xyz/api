package search

import (
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/account"
	"github.com/x-xyz/goapi/domain/collection"
	"github.com/x-xyz/goapi/domain/nftitem"
)

type Result struct {
	Accounts    []*account.SimpleAccount `json:"accounts,omitempty"`
	Collections []*collection.Collection `json:"collections,omitempty"`
	Tokens      []*nftitem.SimpleNftItem `json:"tokens,omitempty"`
}

const (
	Account    = "account"
	Collection = "collection"
	Token      = "token"
)

type Usecase interface {
	Search(c ctx.Ctx, keyword string, filter []string, collections []domain.Address) (*Result, error)
	SearchAccounts(c ctx.Ctx, keyword string) (*Result, error)
	SearchCollections(c ctx.Ctx, keyword string, collections []domain.Address) (*Result, error)
	SearchTokens(c ctx.Ctx, keyword string, collections []domain.Address) (*Result, error)
}
