package hyype

import (
	"errors"
	"net/http"
	"time"

	bCtx "github.com/x-xyz/goapi/base/ctx"
)

var (
	ErrStatusCodeNotOk = errors.New("http.status != 200")
)

type Client interface {
	GetLoresOfNft(bCtx.Ctx, string, string, int, int) ([]byte, error)
}

type ClientCfg struct {
	HttpClient http.Client
	Timeout    time.Duration
	Apikey     string
}
