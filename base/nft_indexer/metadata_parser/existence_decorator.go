package metadata_parser

import (
	"fmt"

	"github.com/x-xyz/goapi/domain/nftitem"

	bCtx "github.com/x-xyz/goapi/base/ctx"
)

type Decorator interface {
	Decorate(ctx bCtx.Ctx, attrs nftitem.Attributes) (nftitem.Attributes, error)
}

type existenceDecorator struct {
	attributes []string
}

func NewExistenceDecorator(attributes []string) Decorator {
	return &existenceDecorator{
		attributes: attributes,
	}
}

func (d *existenceDecorator) Decorate(ctx bCtx.Ctx, attrs nftitem.Attributes) (nftitem.Attributes, error) {
	unprocessedAttrs := make(map[string]interface{})
	for _, attr := range d.attributes {
		unprocessedAttrs[attr] = struct{}{}
	}
	for _, attr := range attrs {
		if _, ok := unprocessedAttrs[attr.TraitType]; ok {
			attrs = append(attrs, toExistenceAttr(attr.TraitType, true))
			delete(unprocessedAttrs, attr.TraitType)
		}
	}
	for attr := range unprocessedAttrs {
		attrs = append(attrs, toExistenceAttr(attr, false))
	}
	return attrs, nil
}

func toExistenceAttr(attrName string, exists bool) nftitem.Attribute {
	existence := "No"
	if exists {
		existence = "Yes"
	}
	return nftitem.Attribute{
		TraitType: fmt.Sprintf("%s?", attrName),
		Value:     existence,
	}
}
