package metadata_parser

import (
	"encoding/json"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/nftitem"
)

type MetadataParser interface {
	Name() string
	Parse(c ctx.Ctx, chainId domain.ChainId, collection domain.Address, tokenId domain.TokenId, data []byte) (nftitem.Attributes, error)
}

type attributesParser struct{}

func NewAttributesParser() MetadataParser {
	return &attributesParser{}
}

func (im *attributesParser) Name() string {
	return "Attributes Parser"
}

func (im *attributesParser) Parse(c ctx.Ctx, _ domain.ChainId, _ domain.Address, _ domain.TokenId, data []byte) (nftitem.Attributes, error) {
	type metadata struct {
		Attributes []nftitem.RawAttribute `json:"attributes"`
	}

	meta := &metadata{}

	if err := json.Unmarshal(data, meta); err != nil {
		return nil, domain.ErrInvalidJsonFormat
	}

	if len(meta.Attributes) == 0 {
		return nil, domain.ErrNotFound
	}

	attrs := nftitem.Attributes{}

	for _, v := range meta.Attributes {
		str, err := json.Marshal(v.Value)
		if err != nil {
			return nil, domain.ErrInvalidJsonFormat
		}

		attrs = append(attrs, nftitem.Attribute{TraitType: v.TraitType, Value: string(str), DisplayType: v.DisplayType})
	}

	return attrs, nil
}

type propertiesParser struct{}

func NewPropertiesParser() MetadataParser {
	return &propertiesParser{}
}

func (im *propertiesParser) Name() string {
	return "Properties Parser"
}

func (im *propertiesParser) Parse(c ctx.Ctx, _ domain.ChainId, _ domain.Address, _ domain.TokenId, data []byte) (nftitem.Attributes, error) {
	type metadata struct {
		Properties nftitem.Properties `json:"properties"`
	}

	meta := &metadata{}

	if err := json.Unmarshal(data, meta); err != nil {
		return nil, domain.ErrInvalidJsonFormat
	}

	if len(meta.Properties) == 0 {
		return nil, domain.ErrNotFound
	}

	attrs := nftitem.Attributes{}

	for k, v := range meta.Properties {
		str, err := json.Marshal(v)
		if err != nil {
			return nil, domain.ErrInvalidJsonFormat
		}

		attrs = append(attrs, nftitem.Attribute{TraitType: k, Value: string(str)})
	}

	return attrs, nil
}

type propertyDetailParser struct{}

func NewPropertyDetailParser() MetadataParser {
	return &propertyDetailParser{}
}

func (im *propertyDetailParser) Name() string {
	return "PropertyDetail Parser"
}

func (im *propertyDetailParser) Parse(c ctx.Ctx, _ domain.ChainId, _ domain.Address, _ domain.TokenId, data []byte) (nftitem.Attributes, error) {
	type metadata struct {
		Properties nftitem.PropertyDetails `json:"properties"`
	}

	meta := &metadata{}

	if err := json.Unmarshal(data, meta); err != nil {
		return nil, domain.ErrInvalidJsonFormat
	}

	if len(meta.Properties) == 0 {
		return nil, domain.ErrNotFound
	}

	attrs := nftitem.Attributes{}

	for _, v := range meta.Properties {
		str, err := json.Marshal(v.Value)
		if err != nil {
			return nil, domain.ErrInvalidJsonFormat
		}

		attrs = append(attrs, nftitem.Attribute{TraitType: v.Name, Value: string(str)})
	}

	return attrs, nil
}
