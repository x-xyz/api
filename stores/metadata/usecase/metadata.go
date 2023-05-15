package usecase

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"

	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain"
	"golang.org/x/xerrors"
)

var ErrInvalidJsonFormat = xerrors.Errorf("invalid JSON form")

type MetadataUseCaseCfg struct {
	CtxTimeout         time.Duration
	HttpReader         domain.WebResourceReaderRepository
	IpfsReader         domain.WebResourceReaderRepository
	DataUriReader      domain.WebResourceReaderRepository
	CloudStorageWriter domain.WebResourceWriterRepository
}

type metadataUseCase struct {
	httpReader         domain.WebResourceReaderRepository
	ipfsReader         domain.WebResourceReaderRepository
	dataUriReader      domain.WebResourceReaderRepository
	cloudStorageWriter domain.WebResourceWriterRepository
}

func NewMetadataUseCase(cfg *MetadataUseCaseCfg) domain.MetadataUseCase {
	return &metadataUseCase{
		httpReader:         cfg.HttpReader,
		ipfsReader:         cfg.IpfsReader,
		dataUriReader:      cfg.DataUriReader,
		cloudStorageWriter: cfg.CloudStorageWriter,
	}
}
func (u *metadataUseCase) GetFromUrl(c bCtx.Ctx, rawUrl string) (*domain.Metadata, error) {
	var (
		data []byte
		err  error
	)

	pUrl, err := url.Parse(rawUrl)
	if err != nil {
		c.WithFields(log.Fields{
			"url": rawUrl,
			"err": err,
		}).Error("failed to parse url")
		return nil, err
	}

	if pUrl.Scheme == "https" {
		data, err = u.httpReader.Get(c, rawUrl)
	} else if pUrl.Scheme == "ipfs" {
		data, err = u.ipfsReader.Get(c, strings.TrimPrefix(rawUrl, "ipfs://"))
	} else if pUrl.Scheme == "data" {
		data, err = u.dataUriReader.Get(c, rawUrl)
	} else {
		return nil, domain.ErrUnsupportedSchema
	}

	if err != nil {
		c.WithFields(log.Fields{
			"schema": pUrl.Scheme,
			"url":    rawUrl,
			"err":    err,
		}).Error("failed to fetch")
		return nil, err
	}
	if !json.Valid(data) {
		c.WithFields(log.Fields{
			"url": rawUrl,
		}).Error("invalid json")
		return nil, ErrInvalidJsonFormat
	}

	return &domain.Metadata{RawMessage: data}, nil
}

func (u *metadataUseCase) Store(c bCtx.Ctx, chainId domain.ChainId, contractAddress string, tokenId int32, metadata *domain.Metadata) (string, error) {
	path := path.Join(
		fmt.Sprintf("%d", chainId),
		contractAddress,
		fmt.Sprintf("%d", tokenId),
	)
	return u.cloudStorageWriter.Store(c, path, metadata.RawMessage, "")
}
