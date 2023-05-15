package file

import (
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/service/pinata"
)

type Usecase interface {
	Upload(c ctx.Ctx, imgData string, pinOption pinata.PinOptions) (hash string, err error)
	UploadJson(c ctx.Ctx, file interface{}, pinOption pinata.PinOptions) (hash string, err error)
}
