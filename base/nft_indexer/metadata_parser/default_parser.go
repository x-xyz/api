package metadata_parser

import (
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/nftitem"

	bCtx "github.com/x-xyz/goapi/base/ctx"
)

type defaultParser struct {
	parsers []MetadataParser
}

func NewDefaultParser() MetadataParser {
	return &defaultParser{
		parsers: []MetadataParser{
			NewAttributesParser(),
			// order of property parser is matter. Property is superset of PropertyDetail.
			// we have to parse PropertyDetail before parsing Property
			NewPropertyDetailParser(),
			NewPropertiesParser(),
		},
	}
}

func (p *defaultParser) Name() string {
	return "Default Parser"
}

func (p *defaultParser) Parse(ctx bCtx.Ctx, chainId domain.ChainId, collection domain.Address, tokenId domain.TokenId, data []byte) (nftitem.Attributes, error) {
	var (
		attrs nftitem.Attributes
		err   error
	)
	for _, parser := range p.parsers {
		attrs, err = parser.Parse(ctx, chainId, collection, tokenId, data)
		if err == nil {
			return attrs, nil
		}
	}
	return nil, err
}
