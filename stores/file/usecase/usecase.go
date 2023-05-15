package usecase

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain/file"
	"github.com/x-xyz/goapi/service/pinata"
)

const (
	imgDataHeaderPrefix    = "data:image/"
	imgDataHeaderSuffix    = ";base64,"
	imgDataHeaderMaxLength = 50
)

type impl struct {
	pinata pinata.Service
}

func New(pinata pinata.Service) file.Usecase {
	return &impl{
		pinata: pinata,
	}
}
func (im *impl) Upload(c ctx.Ctx, imgData string, pinOption pinata.PinOptions) (hash string, err error) {
	reader, extension, err := im.parseImgData(imgData)
	if err != nil {
		c.WithField("err", err).Error("im.parseImgData failed")
		return "", err
	}

	opts := []pinata.Options{}
	if pinOption.Metadata != nil {
		opts = append(opts, pinata.WithMetadata(*pinOption.Metadata))
	}
	if pinOption.Options != nil {
		opts = append(opts, pinata.WithOptions(*pinOption.Options))
	}
	hash, err = im.pinata.Pin(c, reader, extension, opts...)
	if err != nil {
		c.WithField("err", err).Error("im.pinata.Pin failed")
		return "", err
	}
	c.WithField("hash", hash).Info("im.pinata.Pin success")
	return hash, err
}

func (im *impl) UploadJson(c ctx.Ctx, file interface{}, pinOption pinata.PinOptions) (hash string, err error) {
	opts := []pinata.Options{}
	if pinOption.Metadata != nil {
		opts = append(opts, pinata.WithMetadata(*pinOption.Metadata))
	}
	if pinOption.Options != nil {
		opts = append(opts, pinata.WithOptions(*pinOption.Options))
	}
	hash, err = im.pinata.PinJson(c, file, opts...)
	if err != nil {
		c.WithField("err", err).Error("im.pinata.Pin failed")
		return "", err
	}
	c.WithField("hash", hash).Info("im.pinata.Pin success")
	return hash, err
}

func (im *impl) parseImgData(data string) (reader io.Reader, extension string, err error) {
	if !strings.HasPrefix(data, imgDataHeaderPrefix) {
		return nil, "", fmt.Errorf("imeage data has wrong prefix")
	}
	// search header suffix in a limited range
	searchLength := imgDataHeaderMaxLength
	if len(data) < searchLength {
		searchLength = len(data)
	}
	headerSuffixIdx := strings.Index(data[:imgDataHeaderMaxLength], imgDataHeaderSuffix)
	if headerSuffixIdx == -1 {
		return nil, "", fmt.Errorf("can't find image data header suffix")
	}

	extension = data[len(imgDataHeaderPrefix):headerSuffixIdx]
	dataStartIdx := headerSuffixIdx + len(imgDataHeaderSuffix)
	decodedData, err := base64.StdEncoding.DecodeString(data[dataStartIdx:])
	if err != nil {
		return nil, "", err
	}
	return bytes.NewBuffer(decodedData), extension, nil
}
