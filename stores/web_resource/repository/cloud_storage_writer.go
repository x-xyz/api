package repository

import (
	"bytes"
	"io"
	"net/url"
	"time"

	"cloud.google.com/go/storage"
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain"
)

type CloudStorageWriterRepoCfg struct {
	Timeout    time.Duration
	Client     *storage.Client
	BucketName string
	Url        string
}

type cloudStorageWriterRepo struct {
	client     *storage.Client
	bucketName string
	ctxTimeout time.Duration
	baseUrl    *url.URL
}

func NewCloudStorageWriterRepo(cfg *CloudStorageWriterRepoCfg) (domain.WebResourceWriterRepository, error) {
	baseUrl, err := url.Parse(cfg.Url)
	if err != nil {
		return nil, err
	}
	return &cloudStorageWriterRepo{
		client:     cfg.Client,
		bucketName: cfg.BucketName,
		ctxTimeout: cfg.Timeout,
		baseUrl:    baseUrl,
	}, nil
}

func (r *cloudStorageWriterRepo) Store(c bCtx.Ctx, path string, body []byte, contentType string) (string, error) {
	contentPath, err := url.Parse(path)
	if err != nil {
		c.WithFields(log.Fields{
			"path": path,
			"err":  err,
		}).Error("failed to parse path")
		return "", err
	}

	ctx, cancel := bCtx.WithTimeout(c, r.ctxTimeout)
	defer cancel()
	w := r.client.Bucket(r.bucketName).Object(path).NewWriter(ctx)
	if len(contentType) > 0 {
		w.ObjectAttrs.ContentType = contentType
	}
	if _, err := io.Copy(w, bytes.NewReader(body)); err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to copy")
		return "", err
	}
	if err := w.Close(); err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to close writer")
		return "", err
	}
	return r.baseUrl.ResolveReference(contentPath).String(), nil
}
