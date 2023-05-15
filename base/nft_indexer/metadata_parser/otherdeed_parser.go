package metadata_parser

import (
	"fmt"
	"strconv"

	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/nftitem"
)

var (
	kodaApi             = "https://api.otherside.xyz/kodas"
	kodaAttributePrefix = "Koda - "
)

type otherdeedParser struct {
	webResourceUC      domain.WebResourceUseCase
	attributesParser   MetadataParser
	otherdeedDecorator Decorator
	kodaDecorator      Decorator
}

func NewOtherdeedParser(webResourceUC domain.WebResourceUseCase) MetadataParser {
	otherdeedDecorator := NewExistenceDecorator([]string{"Artifact", "Koda"})
	kodaDecorator := NewExistenceDecorator([]string{
		fmt.Sprintf("%sClothing", kodaAttributePrefix),
		fmt.Sprintf("%sWeapon", kodaAttributePrefix),
	})
	return &otherdeedParser{
		webResourceUC:      webResourceUC,
		attributesParser:   NewAttributesParser(),
		otherdeedDecorator: otherdeedDecorator,
		kodaDecorator:      kodaDecorator,
	}
}

func (p *otherdeedParser) Name() string {
	return "Otherdeed Parser"
}

func (p *otherdeedParser) Parse(ctx bCtx.Ctx, chainId domain.ChainId, collection domain.Address, tokenId domain.TokenId, data []byte) (nftitem.Attributes, error) {
	attrs, err := p.attributesParser.Parse(ctx, chainId, collection, tokenId, data)
	if err != nil {
		return nil, err
	}

	attrs, err = p.otherdeedDecorator.Decorate(ctx, attrs)
	if err != nil {
		return nil, err
	}

	kodaId, _hasKoda := hasKoda(attrs)
	if !_hasKoda {
		return attrs, nil
	}

	kodaAttrs, err := p.getKodaAttributes(ctx, chainId, collection, tokenId, kodaId)
	if err != nil {
		return nil, err
	}

	attrs = append(attrs, kodaAttrs...)
	attrs, err = p.kodaDecorator.Decorate(ctx, attrs)
	if err != nil {
		return nil, err
	}
	return attrs, nil
}

func (p *otherdeedParser) getKodaAttributes(ctx bCtx.Ctx, chainId domain.ChainId, collection domain.Address, tokenId domain.TokenId, kodaId string) (nftitem.Attributes, error) {
	url := fmt.Sprintf("%s/%s", kodaApi, kodaId)
	data, err := p.webResourceUC.GetJson(ctx, url)
	if err != nil {
		return nil, err
	}
	attrs, err := p.attributesParser.Parse(ctx, chainId, collection, tokenId, data)
	if err != nil {
		return nil, err
	}
	for i, attr := range attrs {
		attrs[i].TraitType = fmt.Sprintf("%s%s", kodaAttributePrefix, attr.TraitType)
	}

	_isMegaKoda, err := isMegaKoda(kodaId)
	if err != nil {
		return nil, err
	}
	megaKodaAttr := toExistenceAttr(fmt.Sprintf("%s%s", kodaAttributePrefix, "Mega"), _isMegaKoda)
	attrs = append(attrs, megaKodaAttr)
	return attrs, nil
}

func hasKoda(attrs nftitem.Attributes) (string, bool) {
	for _, attr := range attrs {
		if attr.TraitType == "Koda" {
			return attr.Value, true
		}
	}
	return "", false
}

func isMegaKoda(kodaId string) (bool, error) {
	id, err := strconv.ParseInt(kodaId, 10, 32)
	if err != nil {
		return false, err
	}
	// mega: 9901 ~ 9999
	return id > 9900 && id < 10000, nil
}
