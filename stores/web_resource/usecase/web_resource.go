package usecase

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"

	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain"
)

type WebResourceUseCaseCfg struct {
	HttpReader         domain.WebResourceReaderRepository
	IpfsReader         domain.WebResourceReaderRepository
	DataUriReader      domain.WebResourceReaderRepository
	ArUriReader        domain.WebResourceReaderRepository
	CloudStorageWriter domain.WebResourceWriterRepository
}

type webResourceUseCase struct {
	httpReader         domain.WebResourceReaderRepository
	ipfsReader         domain.WebResourceReaderRepository
	dataUriReader      domain.WebResourceReaderRepository
	arUriReader        domain.WebResourceReaderRepository
	cloudStorageWriter domain.WebResourceWriterRepository
}

func NewWebResourceUseCase(cfg *WebResourceUseCaseCfg) domain.WebResourceUseCase {
	return &webResourceUseCase{
		httpReader:         cfg.HttpReader,
		ipfsReader:         cfg.IpfsReader,
		dataUriReader:      cfg.DataUriReader,
		arUriReader:        cfg.ArUriReader,
		cloudStorageWriter: cfg.CloudStorageWriter,
	}
}

func (u *webResourceUseCase) Get(c bCtx.Ctx, rawUrl string) ([]byte, error) {
	return u.get(c, rawUrl)
}

func (u *webResourceUseCase) GetJson(c bCtx.Ctx, rawUrl string) ([]byte, error) {
	data, err := u.get(c, rawUrl)
	if err != nil {
		return nil, err
	}
	if !json.Valid(data) {
		c.WithFields(log.Fields{
			"url": rawUrl,
		}).Error("invalid json")
		return nil, domain.ErrInvalidJsonFormat
	}

	return data, nil
}

func (u *webResourceUseCase) get(c bCtx.Ctx, rawUrl string) ([]byte, error) {
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
	} else if pUrl.Scheme == "http" {
		data, err = u.httpReader.Get(c, rawUrl)
	} else if pUrl.Scheme == "ipfs" {
		ipfsUrl := strings.TrimPrefix(rawUrl, "ipfs://")
		ipfsUrl = strings.TrimPrefix(ipfsUrl, "ipfs/") // early foundation's metadata bug
		data, err = u.ipfsReader.Get(c, ipfsUrl)
	} else if pUrl.Scheme == "data" {
		data, err = u.dataUriReader.Get(c, rawUrl)
	} else if pUrl.Scheme == "ar" {
		data, err = u.arUriReader.Get(c, rawUrl)
	} else {
		return nil, domain.ErrUnsupportedSchema
	}

	if err == nil {
		return data, nil
	}

	if pUrl.Scheme == "https" {
		ipfsUrl := getIpfsUrl(rawUrl)
		if len(ipfsUrl) > 0 {
			c.WithFields(log.Fields{
				"url":     rawUrl,
				"ipfsUrl": ipfsUrl,
			}).Info("falling back to ipfs")
			return u.get(c, ipfsUrl)
		}
	}

	c.WithFields(log.Fields{
		"schema": pUrl.Scheme,
		"url":    rawUrl,
		"err":    err,
	}).Error("failed to fetch")
	return nil, err
}

func (u *webResourceUseCase) Store(c bCtx.Ctx, chainId domain.ChainId, contractAddress domain.Address, tokenId domain.TokenId, _typ string, ext string, data []byte, contentType string) (string, error) {
	_path := path.Join(
		fmt.Sprintf("%d", chainId),
		contractAddress.ToLowerStr(),
		fmt.Sprintf("%s.%s%s", tokenId, _typ, ext),
	)
	url, err := u.cloudStorageWriter.Store(c, _path, data, contentType)
	if err != nil {
		c.WithFields(log.Fields{
			"path": _path,
			"err":  err,
		}).Error("cloudStorageWriter.Store failed")
		return "", err
	}
	return url, nil
}

func getIpfsUrl(url string) string {
	var (
		pinataPrefix     = "https://gateway.pinata.cloud/ipfs/"
		ipfsIoPrefix     = "https://ipfs.io/ipfs/"
		cloudflarePrefix = "https://cloudflare-ipfs.com/ipfs/"
		foundationPrefix = "https://ipfs.foundation.app/ipfs/"
		ipfsPrefix       = "ipfs://"
	)

	fixedPrefix := []string{pinataPrefix, ipfsIoPrefix, cloudflarePrefix, foundationPrefix}
	for _, p := range fixedPrefix {
		if strings.HasPrefix(url, p) {
			return strings.Replace(url, p, ipfsPrefix, 1)
		}
	}
	dedicatedPinataRegex := regexp.MustCompile(`^https://.*.mypinata.cloud/ipfs/`)
	if dedicatedPinataRegex.Match([]byte(url)) {
		return dedicatedPinataRegex.ReplaceAllLiteralString(url, ipfsPrefix)
	}
	return ""
}
