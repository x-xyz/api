package pinata

import (
	"errors"
	"io"

	"github.com/x-xyz/goapi/base/ctx"
)

var (
	ErrRequestFailed = errors.New("request failed")
)

type PinataMetadata struct {
	Name string `json:"name,omitempty"`
	// can only store string, bool, int
	KeyValues map[string]interface{} `json:"keyvalues,omitempty"`
}

type PinataOptions struct {
	CidVersion CidVersion `json:"cidVersion"`
}

type CidVersion uint8

const (
	CidVersion_0 CidVersion = 0
	CidVersion_1 CidVersion = 1
)

type PinOptions struct {
	Metadata      *PinataMetadata `json:"pinataMetadata"`
	Options       *PinataOptions  `json:"pinataOptions"`
	PinataContent interface{}     `json:"pinataContent"`
}

type Options func(*PinOptions) error

func GetPinOptions(opts ...Options) (*PinOptions, error) {
	res := &PinOptions{}

	for _, opt := range opts {
		if err := opt(res); err != nil {
			return nil, err
		}
	}

	return res, nil
}

func WithMetadata(metadata PinataMetadata) Options {
	return func(options *PinOptions) error {
		options.Metadata = &metadata
		return nil
	}
}

func WithOptions(pinataOptions PinataOptions) Options {
	return func(options *PinOptions) error {
		options.Options = &pinataOptions
		return nil
	}
}

type Service interface {
	Pin(c ctx.Ctx, file io.Reader, extension string, opts ...Options) (string, error)
	PinJson(c ctx.Ctx, value interface{}, opts ...Options) (string, error)
}
