package domain

import (
	"github.com/x-xyz/goapi/base/ctx"
)

type OpenseaData struct {
	ChainId           ChainId `bson:"chainId" json:"chainId"`
	Address           Address `bson:"address" json:"address"`
	Slug              string  `bson:"slug" json:"slug"`
	Name              string  `bson:"name" json:"name"`
	Description       string  `bson:"description" json:"description"`
	ImageUrl          string  `bson:"imageUrl" json:"imageUrl"`
	ExternalUrl       string  `bson:"externalUrl" json:"externalUrl"`
	DiscordUrl        string  `bson:"discordUrl" json:"discordUrl"`
	TwitterUserName   string  `bson:"twitterUsername" json:"twitterUsername"`
	InstagramUserName string  `bson:"instagramUsername" json:"instagramUsername"`
	MediumUserName    string  `bson:"mediumUsername" json:"mediumUsername"`
	TelegramUrl       string  `bson:"telegramUrl" json:"telegramUrl"`
	PayoutAddress     Address `bson:"payoutAddress" json:"payoutAddress"`
	Royalty           uint16  `bson:"royalty" json:"royalty"`
	Symbol            string  `bson:"symbol" json:"symbol"`
	OneHourVolume     float64 `bson:"oneHourVolume" json:"oneHourVolume"`
	OneHourChange     float64 `bson:"oneHourChange" json:"oneHourChange"`
	OneHourSales      float64 `bson:"oneHourSales" json:"oneHourSales"`
	SixHourVolume     float64 `bson:"sixHourVolume" json:"sixHourVolume"`
	SixHourChange     float64 `bson:"sixHourChange" json:"sixHourChange"`
	SixHourSales      float64 `bson:"sixHourSales" json:"sixHourSales"`
	OneDayVolume      float64 `bson:"oneDayVolume" json:"oneDayVolume"`
	OneDayChange      float64 `bson:"oneDayChange" json:"oneDayChange"`
	OneDaySales       float64 `bson:"oneDaySales" json:"oneDaySales"`
	SevenDayVolume    float64 `bson:"sevenDayVolume" json:"sevenDayVolume"`
	SevenDayChange    float64 `bson:"sevenDayChange" json:"sevenDayChange"`
	SevenDaySales     float64 `bson:"sevenDaySales" json:"sevenDaySales"`
	ThirtyDayVolume   float64 `bson:"thirtyDayVolume" json:"thirtyDayVolume"`
	ThirtyDayChange   float64 `bson:"thirtyDayChange" json:"thirtyDayChange"`
	ThirtyDaySales    float64 `bson:"thirtyDaySales" json:"thirtyDaySales"`
	TotalVolume       float64 `bson:"totalVolume" json:"totalVolume"`
	TotalSales        float64 `bson:"totalSales" json:"totalSales"`
	FloorPrice        float64 `bson:"floorPrice" json:"floorPrice"`
}

func (d OpenseaData) ToId() OpenseaDataId {
	return OpenseaDataId{
		ChainId: d.ChainId,
		Address: d.Address,
	}
}

type OpenseaDataId struct {
	ChainId ChainId `bson:"chainId" json:"chainId"`
	Address Address `bson:"address" json:"address"`
}

type openseaDataFindAllOptions struct {
	SortBy    *string    `bson:"-"`
	SortDir   *SortDir   `bson:"-"`
	Offset    *int32     `bson:"-"`
	Limit     *int32     `bson:"-"`
	ChainId   *ChainId   `bson:"chainId"`
	Addresses *[]Address `bson:"-"`
}
type OpenseaDataFindAllOptions func(*openseaDataFindAllOptions) error

func GetOpenseaDataFindAllOptions(opts ...OpenseaDataFindAllOptions) (openseaDataFindAllOptions, error) {
	res := openseaDataFindAllOptions{}

	for _, opt := range opts {
		if err := opt(&res); err != nil {
			return res, err
		}
	}

	return res, nil
}

func OpenseaDataWithSort(sortby string, sortdir SortDir) OpenseaDataFindAllOptions {
	return func(options *openseaDataFindAllOptions) error {
		options.SortBy = &sortby
		options.SortDir = &sortdir
		return nil
	}
}

func OpenseaDataWithPagination(offset int32, limit int32) OpenseaDataFindAllOptions {
	return func(options *openseaDataFindAllOptions) error {
		options.Offset = &offset
		options.Limit = &limit
		return nil
	}
}

func OpenseaDataWithChainId(chainId ChainId) OpenseaDataFindAllOptions {
	return func(options *openseaDataFindAllOptions) error {
		options.ChainId = &chainId
		return nil
	}
}

var YugaLabCollectionAddresses = []Address{
	"0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
	"0x60e4d786628fea6478f785a6d7e704777c86a7c6",
	"0x34d85c9cdeb23fa97cb08333b511ac86e1c4e258",
	"0xba30e5f9bb24caa003e9f2f0497ad287fdf95623",
	"0x22c36bfdcef207f9c0cc941936eff94d4246d14a",
	"0xb7f7f6c52f2e2fdb1963eab30438024864c313f6",
	"0x7bd29408f11d2bfc23c34f18275bbf23bb716bc7",
	"0x1a92f7381b9f03921564a437210bb9396471050c",
	"0x1cb1a5e65610aeff2551a50f76a87a7d3fb649c6",
	"0x9c8ff314c9bc7f6e59a9d9225fb22946427edc03",
	"0xe785e82358879f061bc3dcac6f0444462d4b5330",
	"0xf61f24c2d93bf2de187546b14425bf631f28d6dc",
	"0x09f717f77b5e7f2d2f37604fec3d0e3d53eb9808",
	"0x0cfb5d82be2b949e8fa73a656df91821e2ad99fd",
	"0x572e33ffa523865791ab1c26b42a86ac244df784",
	"0xccc1825cf04cae4d497b202d1434ec0f79ee535f",
	"0x7daec605e9e2a1717326eedfd660601e2753a057",
	"0x19445bb9f1b704dd973f8f9a4dce5ea5e55444da",
	"0xae99a698156ee8f8d07cbe7f271c31eeaac07087",
	"0x5bd13ff0279639f7c27da270e4a0e1a73f073de8",
	"0x354634c4621cdfb7a25e6486cca1e019777d841b",
	"0x89c3df79aa8a3cbc96caf32f83eba8f1bd3787b9",
	"0xba627f3d081cc97ac0edc40591eda7053ac63532",
	"0x880644ddf208e471c6f2230d31f9027578fa6fcc",
	"0x85ed8f10c4889b4bc60400a0c1f796254d35003d",
	"0xe75ef1ec029c71c9db0f968e389331609312aa22",
	"0x764aeebcf425d56800ef2c84f2578689415a2daa",
	"0x4b15a9c28034dc83db40cd810001427d3bd7163d",
	"0x790b2cf29ed4f310bf7641f013c65d4560d28371",
	"0xe012baf811cf9c05c408e879c399960d1f305903",
	"0x5b1085136a811e55b2bb2ca1ea456ba82126a376",
	"0x0dfc1bc020e2d6a7cf234894e79686a88fbe2b2a",
	"0xb7abcc333209f3e8af129ab99ee2a470a2ac5bae",
}

var MetadataRefreshCollectionAddresses = []Address{
	// "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d", // BAYC
	// "0x60e4d786628fea6478f785a6d7e704777c86a7c6", // MAYC
	// "0x34d85c9cdeb23fa97cb08333b511ac86e1c4e258", // Otherdeed
	// "0xba30e5f9bb24caa003e9f2f0497ad287fdf95623", // BAKC
	// "0x22c36bfdcef207f9c0cc941936eff94d4246d14a",
	// "0xb7f7f6c52f2e2fdb1963eab30438024864c313f6",
	// "0x7bd29408f11d2bfc23c34f18275bbf23bb716bc7",
	// "0x1a92f7381b9f03921564a437210bb9396471050c",
	// "0x1cb1a5e65610aeff2551a50f76a87a7d3fb649c6",
	// "0x9c8ff314c9bc7f6e59a9d9225fb22946427edc03",
	// "0xe785e82358879f061bc3dcac6f0444462d4b5330",
	// "0xf61f24c2d93bf2de187546b14425bf631f28d6dc",
	// "0x09f717f77b5e7f2d2f37604fec3d0e3d53eb9808",
	// "0x0cfb5d82be2b949e8fa73a656df91821e2ad99fd",
	// "0x572e33ffa523865791ab1c26b42a86ac244df784",
	// "0xccc1825cf04cae4d497b202d1434ec0f79ee535f",
	// "0x7daec605e9e2a1717326eedfd660601e2753a057",
	// "0x19445bb9f1b704dd973f8f9a4dce5ea5e55444da",
	// "0xae99a698156ee8f8d07cbe7f271c31eeaac07087",
	// "0x5bd13ff0279639f7c27da270e4a0e1a73f073de8",
	// "0x354634c4621cdfb7a25e6486cca1e019777d841b",
	// "0x89c3df79aa8a3cbc96caf32f83eba8f1bd3787b9",
	// "0xba627f3d081cc97ac0edc40591eda7053ac63532",
	// "0x880644ddf208e471c6f2230d31f9027578fa6fcc",
	// "0x85ed8f10c4889b4bc60400a0c1f796254d35003d",
	// "0xe75ef1ec029c71c9db0f968e389331609312aa22",
	"0x764aeebcf425d56800ef2c84f2578689415a2daa", // sewer pass
	"0x4b15a9c28034dc83db40cd810001427d3bd7163d", // HV-MTL
}

func OpenseaDataWithAddresses(addresses []Address) OpenseaDataFindAllOptions {
	return func(options *openseaDataFindAllOptions) error {
		options.Addresses = &addresses
		return nil
	}
}

type OpenseaDataUseCase interface {
	FindOne(ctx.Ctx, OpenseaDataId) (*OpenseaData, error)
	Upsert(ctx.Ctx, OpenseaData) error
}

type OpenseaDataRepo interface {
	FindAll(c ctx.Ctx, opts ...OpenseaDataFindAllOptions) ([]OpenseaData, error)
	FindOne(ctx.Ctx, OpenseaDataId) (*OpenseaData, error)
	Upsert(ctx.Ctx, OpenseaData) error
}
