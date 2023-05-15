package metadata_parser

import (
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/apecoinstaking"
	"github.com/x-xyz/goapi/domain/nftitem"
)

type baycParser struct {
	attributesParser MetadataParser
	apecoinStakingUC apecoinstaking.UseCase
}

func NewBAYCParser(apecoinStakingUC apecoinstaking.UseCase) MetadataParser {
	return &baycParser{
		attributesParser: NewAttributesParser(),
		apecoinStakingUC: apecoinStakingUC,
	}
}

func (p *baycParser) Name() string {
	return "BAYC Parser"
}

func (p *baycParser) Parse(ctx bCtx.Ctx, chainId domain.ChainId, collection domain.Address, tokenId domain.TokenId, data []byte) (nftitem.Attributes, error) {
	attrs, err := p.attributesParser.Parse(ctx, chainId, collection, tokenId, data)
	if err != nil {
		return nil, err
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
