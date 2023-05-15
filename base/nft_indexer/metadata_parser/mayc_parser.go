package metadata_parser

import (
	"strings"

	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/apecoinstaking"
	"github.com/x-xyz/goapi/domain/nftitem"
)

var (
	mutationTypeTraitName = "Mutation Type"
	m1TraitValue          = "M1"
	m2TraitValue          = "M2"
	megaTraitValue        = "Mega"
)

type maycParser struct {
	attributesParser MetadataParser
	apecoinStakingUC apecoinstaking.UseCase
}

func NewMAYCParser(apecoinStakingUC apecoinstaking.UseCase) MetadataParser {
	return &maycParser{
		attributesParser: NewAttributesParser(),
		apecoinStakingUC: apecoinStakingUC,
	}
}

func (p *maycParser) Name() string {
	return "MAYC Parser"
}

func (p *maycParser) Parse(ctx bCtx.Ctx, chainId domain.ChainId, collection domain.Address, tokenId domain.TokenId, data []byte) (nftitem.Attributes, error) {
	attrs, err := p.attributesParser.Parse(ctx, chainId, collection, tokenId, data)
	if err != nil {
		return nil, err
	}

	if len(attrs) > 0 {
		if strings.HasPrefix(strings.TrimPrefix(attrs[0].Value, "\""), "M1") {
			attrs = append(attrs, nftitem.Attribute{TraitType: mutationTypeTraitName, Value: m1TraitValue})
		} else if strings.HasPrefix(strings.TrimPrefix(attrs[0].Value, "\""), "M2") {
			attrs = append(attrs, nftitem.Attribute{TraitType: mutationTypeTraitName, Value: m2TraitValue})
		} else {
			attrs = append(attrs, nftitem.Attribute{TraitType: mutationTypeTraitName, Value: megaTraitValue})
		}
	}

	id := apecoinstaking.Id{
		ChainId:         chainId,
		ContractAddress: collection,
		TokenId:         tokenId,
	}
	staking, err := p.apecoinStakingUC.Get(ctx, id)
	if err != nil && err != domain.ErrNotFound {
		ctx.WithFields(log.Fields{"id": id, "err": err}).Error("apecoinStakingUC.Get failed")
		return nil, err
	}
	if err == nil && staking.Staked {
		attrs = append(attrs, nftitem.Attribute{TraitType: stakedTraitType, Value: "Yes"})
	} else {
		attrs = append(attrs, nftitem.Attribute{TraitType: stakedTraitType, Value: "No"})
	}

	return attrs, nil
}
