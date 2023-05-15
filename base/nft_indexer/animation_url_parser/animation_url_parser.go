package animation_url_parser

import (
	"encoding/json"

	bCtx "github.com/x-xyz/goapi/base/ctx"
)

type Parser interface {
	GetAnimationUrl(bCtx.Ctx, []byte, string) (string, error)
}

type defaultMetadataStruct struct {
	AnimationUrl string `json:"animation_url"`
}

type defaultParser struct {
}

func NewDefaultParser() Parser {
	return &defaultParser{}
}

func (p *defaultParser) GetAnimationUrl(ctx bCtx.Ctx, data []byte, _ string) (string, error) {
	var metadata defaultMetadataStruct
	err := json.Unmarshal(data, &metadata)
	if err != nil {
		ctx.WithField("err", err).Error("json.Unmarshal failed")
		return "", err
	}
	return metadata.AnimationUrl, nil
}
