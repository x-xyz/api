package repository

import (
	"encoding/base64"
	"strings"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
	"golang.org/x/xerrors"
)

const dataUriSchema = "data:"

type dataUriReaderRepo struct {
}

func NewDataUriReaderRepo() domain.WebResourceReaderRepository {
	return &dataUriReaderRepo{}
}

func (r *dataUriReaderRepo) Get(_ ctx.Ctx, uri string) ([]byte, error) {
	if !strings.HasPrefix(uri, dataUriSchema) {
		return nil, xerrors.Errorf("invalid data uri")
	}
	// data:[<mediatype>][;base64],<data>
	uriParts := strings.SplitN(strings.TrimPrefix(uri, dataUriSchema), ",", 2)
	if len(uriParts) < 2 || len(uriParts[1]) == 0 {
		return nil, xerrors.Errorf("no data part provided")
	}

	if strings.HasSuffix(uriParts[0], ";base64") {
		return base64.StdEncoding.DecodeString(uriParts[1])
	}
	// treat as plain text
	return []byte(uriParts[1]), nil
}
