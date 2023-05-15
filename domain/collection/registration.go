package collection

import (
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
)

type RegistrationState int

const (
	RegistrationStatePending RegistrationState = iota
	RegistrationStateAccept
	RegistrationStateReject
)

type Registration struct {
	ChainId         domain.ChainId    `json:"chainId" bson:"chainId"`
	Erc721Address   domain.Address    `json:"erc721Address" bson:"erc721Address"`
	TokenType       domain.TokenType  `json:"tokenType" bson:"tokenType"`
	Owner           domain.Address    `json:"owner" bson:"owner"`
	Email           string            `json:"email" bson:"email"`
	CollectionName  string            `json:"collectionName" bson:"collectionName"`
	Description     string            `json:"description" bson:"description"`
	Categories      []string          `json:"categories" bson:"categories"`
	LogoImage       string            `json:"logoImage" bson:"-"`
	LogoImageHash   string            `json:"-" bson:"logoImageHash"`
	LogoImageUrl    string            `json:"logoImageUrl" bson:"logoImageUrl"`
	CoverImage      string            `json:"coverImage" bson:"-"`
	CoverImageHash  string            `json:"-" bson:"coverImageHash"`
	CoverImageURL   string            `json:"coverImageUrl" bson:"coverImageUrl"`
	SiteUrl         string            `json:"siteUrl" bson:"siteUrl"`
	Discord         string            `json:"discord" bson:"discord"`
	TwitterHandle   string            `json:"twitterHandle" bson:"twitterHandle"`
	InstagramHandle string            `json:"instagramHandle" bson:"instagramHandle"`
	MediumHandle    string            `json:"mediumHandle" bson:"mediumHandle"`
	Telegram        string            `json:"telegram" bson:"telegram"`
	Royalty         float64           `json:"royalty" bson:"royalty"`
	FeeRecipient    string            `json:"feeRecipient" bson:"feeRecipient"`
	State           RegistrationState `json:"-" bson:"state"`
}

type UpdateRegistration struct {
	State RegistrationState `json:"-" bson:"state"`
}

type RegistrationRepo interface {
	FindAll(c ctx.Ctx) ([]*Registration, error)
	FindOne(c ctx.Ctx, id CollectionId) (*Registration, error)
	Create(c ctx.Ctx, value Registration) error
	Patch(c ctx.Ctx, id CollectionId, value UpdateRegistration) error
}
