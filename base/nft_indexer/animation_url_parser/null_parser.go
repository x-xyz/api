package animation_url_parser

import (
	bCtx "github.com/x-xyz/goapi/base/ctx"
)

type nullParser struct {
}

func NewNullParser() Parser {
	return &nullParser{}
}

func (p *nullParser) GetAnimationUrl(ctx bCtx.Ctx, data []byte, _ string) (string, error) {
	return "", nil
}
