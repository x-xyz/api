package nft_indexer

import (
	"encoding/json"
	"errors"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gabriel-vasile/mimetype"
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/base/metrics"
	"github.com/x-xyz/goapi/base/nft_indexer/animation_url_parser"
	"github.com/x-xyz/goapi/base/nft_indexer/metadata_parser"
	"github.com/x-xyz/goapi/base/ptr"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/collection"
	"github.com/x-xyz/goapi/domain/nftitem"
	"github.com/x-xyz/goapi/domain/token"
)

var (
	met     metrics.Service
	metOnce sync.Once

	errBadDataUriFormat = errors.New("data uri format not correct")
)

type metadataStruct struct {
	Image        string `json:"image"`
	ImageUrl     string `json:"image_url"`
	ImageData    string `json:"image_data"`
	AnimationURL string `json:"animation_url"`
	Name         string `json:"name"`
}

type NftIndexerCfg struct {
	WebResourceUC              domain.WebResourceUseCase
	TokenUC                    token.Usecase
	TargetState                nftitem.IndexerState
	RetryLimit                 int
	Batch                      int
	Workers                    int
	Interval                   time.Duration
	ErrorCh                    chan<- error
	ThumborUrl                 string
	ParserSelector             *metadata_parser.Selector
	AnimationUrlParserSelector *animation_url_parser.Selector
}

type NftIndexer struct {
	webResourceUC domain.WebResourceUseCase
	tokenUC       token.Usecase
	targetState   nftitem.IndexerState
	retryLimit    int
	batch         int
	workers       int
	interval      time.Duration
	taskCh        chan *nftitem.NftItem
	errorCh       chan<- error
	stoppedCh     chan interface{}

	thumborUrl                 string
	parserSelector             *metadata_parser.Selector
	animationUrlParserSelector *animation_url_parser.Selector
}

func NewNftIndexer(cfg *NftIndexerCfg) *NftIndexer {
	metOnce.Do(func() {
		met = metrics.New("indexer")
	})
	return &NftIndexer{
		webResourceUC: cfg.WebResourceUC,
		tokenUC:       cfg.TokenUC,
		targetState:   cfg.TargetState,
		retryLimit:    cfg.RetryLimit,
		batch:         cfg.Batch,
		workers:       cfg.Workers,
		interval:      cfg.Interval,
		taskCh:        make(chan *nftitem.NftItem, cfg.Batch),
		stoppedCh:     make(chan interface{}),

		thumborUrl:                 cfg.ThumborUrl,
		parserSelector:             cfg.ParserSelector,
		animationUrlParserSelector: cfg.AnimationUrlParserSelector,
	}
}

func (i *NftIndexer) Start(ctx bCtx.Ctx) {
	go i.loop(ctx)
}

func (i *NftIndexer) Wait() {
	<-i.stoppedCh
}

func (i *NftIndexer) loop(ctx bCtx.Ctx) {
	workerCtx, cancel := bCtx.WithCancel(ctx)
	workerWg := sync.WaitGroup{}
	nextTick := time.Second * 0
	resCh := make(chan error, i.workers)

	errAndStop := func(err error) {
		i.errorCh <- err
		cancel()
		workerWg.Wait()
		close(i.stoppedCh)
	}

	for j := 0; j < i.workers; j++ {
		workerWg.Add(1)
		go func() {
			defer workerWg.Done()
			for {
				select {
				case <-workerCtx.Done():
					return
				case task := <-i.taskCh:
					err := i.processNft(workerCtx, task)
					if err != nil {
						resCh <- err
						return
					}
					resCh <- nil
				}
			}
		}()
	}

	for {
		select {
		case <-ctx.Done():
			workerWg.Wait()
			close(i.stoppedCh)
			return
		case <-time.After(nextTick):
			items, count, err := i.tokenUC.SearchForIndexerState(ctx, []nftitem.IndexerState{i.targetState}, i.retryLimit,
				token.WithSort("indexerRetryCount", domain.SortDirAsc),
				token.WithPagination(0, int32(i.batch)),
			)
			if err != nil {
				errAndStop(err)
				return
			}

			ctx.WithFields(log.Fields{
				"#items": len(items),
				"state":  i.targetState,
			}).Info("search result")
			met.BumpAvg("tokenState.count", float64(count), "state", string(i.targetState))
			for _, item := range items {
				i.taskCh <- item
			}
			for j := 0; j < len(items); j++ {
				select {
				case <-ctx.Done():
					cancel()
					// TODO: drain resCh?
					workerWg.Wait()
					close(i.stoppedCh)
					return
				case err := <-resCh:
					if err != nil {
						errAndStop(err)
						return
					}
				}
			}
			if len(items) < i.batch {
				nextTick = i.interval
			} else {
				nextTick = time.Second * 0
			}
		}
	}
}

func (i *NftIndexer) processNft(ctx bCtx.Ctx, item *nftitem.NftItem) error {
	ctx = bCtx.WithValues(ctx, map[string]interface{}{
		"chainId":  item.ChainId,
		"contract": item.ContractAddress,
		"tokenId":  item.TokenId,
		"tokenURI": item.TokenUri,
		"imageUrl": item.ImageUrl,
	})

	switch item.IndexerState {
	case nftitem.IndexerStateHasTokenURI:
		return i.storeMetaAndGetImageURI(ctx, item)
	case nftitem.IndexerStateHasTokenURIRefreshing:
		return i.storeMetaAndGetImageURI(ctx, item)
	case nftitem.IndexerStateHasImageURL:
		return i.storeImage(ctx, item)
	case nftitem.IndexerStateHasHostedImage:
		return i.storeGeneratedThumbnail(ctx, item)
	case nftitem.IndexerStateParsingAttributes:
		return i.storeAttributes(ctx, item)
	case nftitem.IndexerStateFetchingAnimation:
		return i.storeAnimationUrl(ctx, item)
	case nftitem.IndexerStateDone:
		return nil
	case nftitem.IndexerStateBeforeMigrate:
		return i.storeMetaAndGetImageURI(ctx, item)
	case nftitem.IndexerStateBeforeMigrateMimeType:
		return i.migrateMimeType(ctx, item)
	default:
		return nil
	}
}

func (i *NftIndexer) migrateMimeType(ctx bCtx.Ctx, item *nftitem.NftItem) error {
	data, err := i.webResourceUC.Get(ctx, item.ImageUrl)
	if err != nil {
		ctx.WithField("err", err).Error("webresource.Get failed")
		return i.increaseRetryCount(ctx, item)
	}

	// detect mime type
	mtype := mimetype.Detect(data)
	mimeType := mtype.String()
	patchable := &nftitem.PatchableNftItem{
		MimeType:          &mimeType,
		IndexerRetryCount: ptr.Int32(0),
		IndexerState:      (*nftitem.IndexerState)(ptr.String(nftitem.IndexerStateDone)),
	}
	if i.tokenUC.PatchNft(ctx, item.ToId(), patchable); err != nil {
		ctx.WithField("err", err).Error("token.PatchNft failed")
		return i.increaseRetryCount(ctx, item)
	}
	return nil
}

func (i *NftIndexer) storeMetaAndGetImageURI(ctx bCtx.Ctx, item *nftitem.NftItem) error {
	// webresource related failures are not considered errors, only db failures are
	ctx.Info("storeMetaAndGetImageURI")
	data, err := i.webResourceUC.GetJson(ctx, item.TokenUri)
	if err != nil {
		ctx.WithField("err", err).Error("webresource.GetJson failed")
		return i.increaseRetryCount(ctx, item)
	}
	path, err := i.webResourceUC.Store(
		ctx, item.ChainId, item.ContractAddress, item.TokenId,
		"metadata", ".json", data, "",
	)
	if err != nil {
		ctx.WithField("err", err).Error("webresource.Store failed")
		return i.increaseRetryCount(ctx, item)
	}
	var metadata metadataStruct
	err = json.Unmarshal(data, &metadata)
	if err != nil {
		ctx.WithField("err", err).Error("json.Unmarshal failed")
		return i.increaseRetryCount(ctx, item)
	}
	patchable := &nftitem.PatchableNftItem{
		HostedTokenUri:    ptr.String(path),
		IndexerRetryCount: ptr.Int32(0),
	}
	if len(metadata.Name) == 0 {
		patchable.Name = ptr.String(item.TokenId.String())
	} else {
		patchable.Name = ptr.String(metadata.Name)
	}
	if len(metadata.Image) != 0 {
		patchable.ImageUrl = ptr.String(metadata.Image)
		patchable.IndexerState = (*nftitem.IndexerState)(ptr.String(nftitem.IndexerStateHasImageURL))
	} else if len(metadata.ImageUrl) != 0 {
		patchable.ImageUrl = ptr.String(metadata.ImageUrl)
		patchable.IndexerState = (*nftitem.IndexerState)(ptr.String(nftitem.IndexerStateHasImageURL))
	} else if len(metadata.ImageData) != 0 {
		patchable.ImageUrl = ptr.String(metadata.ImageData)
		patchable.IndexerState = (*nftitem.IndexerState)(ptr.String(nftitem.IndexerStateHasImageURL))
	} else if len(metadata.AnimationURL) != 0 {
		patchable.ImageUrl = ptr.String(metadata.AnimationURL)
		patchable.IndexerState = (*nftitem.IndexerState)(ptr.String(nftitem.IndexerStateHasImageURL))
	} else {
		patchable.IndexerState = (*nftitem.IndexerState)(ptr.String(nftitem.IndexerStateInvalid))
	}
	if err := i.tokenUC.PatchNft(ctx, item.ToId(), patchable); err != nil {
		ctx.WithField("err", err).Error("token.PatchNft failed")
		return err
	}
	return nil
}

func extractMimeTypeFromDataUri(s string) (string, error) {
	if !strings.Contains(s, ",") {
		return "", errBadDataUriFormat
	}

	parts := strings.Split(s, ",")
	start := strings.Index(s, ":")
	end := len(parts[0])
	if strings.Contains(s, ";") {
		end = strings.Index(s, ";")
	}

	if start == -1 {
		return "", errBadDataUriFormat
	}

	return s[start+1 : end], nil
}

func (i *NftIndexer) storeImage(ctx bCtx.Ctx, item *nftitem.NftItem) error {
	// webresource related failures are not considered errors, only db failures are
	// ctx.Info("storeImage")
	data, err := i.webResourceUC.Get(ctx, item.ImageUrl)
	if err != nil {
		ctx.WithField("err", err).Error("webresource.Get failed")
		return i.increaseRetryCount(ctx, item)
	}

	// detect mime type
	mtype := mimetype.Detect(data)
	mimeType := mtype.String()
	ext := mtype.Extension()

	path, err := i.webResourceUC.Store(
		ctx, item.ChainId, item.ContractAddress, item.TokenId,
		"media", ext, data, mimeType,
	)
	if err != nil {
		ctx.WithField("err", err).Error("webresource.Store failed")
		return i.increaseRetryCount(ctx, item)
	}
	contentType := mimeToContentType(mimeType)
	patchable := &nftitem.PatchableNftItem{
		ContentType:       &contentType,
		MimeType:          &mimeType,
		ImageUrl:          ptr.String(path),
		HostedImageUrl:    ptr.String(path),
		IndexerState:      (*nftitem.IndexerState)(ptr.String(nftitem.IndexerStateHasHostedImage)),
		IndexerRetryCount: ptr.Int32(0),
	}
	if err := i.tokenUC.PatchNft(ctx, item.ToId(), patchable); err != nil {
		ctx.WithField("err", err).Error("token.PatchNft failed")
		return err
	}
	return nil
}

func (i *NftIndexer) storeGeneratedThumbnail(ctx bCtx.Ctx, item *nftitem.NftItem) error {
	ctx.Info("storeGeneratedThumbnail")

	thumbnailPath := ""
	if filepath.Ext(item.ImageUrl) == ".svg" {
		// svg should not compress to other format
		thumbnailPath = item.ImageUrl
	} else {
		p := i.thumborUrl + path.Join(
			"/unsafe/300x0/filters:format(webp)",
			url.QueryEscape(item.ImageUrl),
		)

		data, err := i.webResourceUC.Get(ctx, p)
		if err != nil {
			ctx.WithField("err", err).Error("webResourceUC.Get failed")
			return i.increaseRetryCount(ctx, item)
		}

		// detect mime type
		mtype := mimetype.Detect(data)
		mimeType := mtype.String()
		ext := mtype.Extension()

		thumbnailPath, err = i.webResourceUC.Store(
			ctx, item.ChainId, item.ContractAddress, item.TokenId,
			"thumbnail", ext, data, mimeType,
		)
		if err != nil {
			ctx.WithField("err", err).Error("webresource.Store failed")
			return i.increaseRetryCount(ctx, item)
		}
	}

	patchable := &nftitem.PatchableNftItem{
		ThumbnailPath:     ptr.String(thumbnailPath),
		IndexerState:      (*nftitem.IndexerState)(ptr.String(nftitem.IndexerStateParsingAttributes)),
		IndexerRetryCount: ptr.Int32(0),
	}
	if err := i.tokenUC.PatchNft(ctx, item.ToId(), patchable); err != nil {
		ctx.WithField("err", err).Error("token.PatchNft failed")
		return err
	}
	return nil
}

func (i *NftIndexer) storeAttributes(ctx bCtx.Ctx, item *nftitem.NftItem) error {
	uri := item.TokenUri
	if len(item.HostedTokenUri) > 0 {
		uri = item.HostedTokenUri
	}
	data, err := i.webResourceUC.GetJson(ctx, uri)
	if err != nil {
		ctx.WithField("err", err).Error("webresource.GetJson failed")
		return i.increaseRetryCount(ctx, item)
	}

	collectionId := collection.CollectionId{ChainId: item.ChainId, Address: item.ContractAddress}
	parser := i.parserSelector.GetParser(collectionId)
	attrs, err := parser.Parse(ctx, item.ChainId, item.ContractAddress, item.TokenId, data)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":        err,
			"parserName": parser.Name(),
		}).Error("parser.Parse failed")
		return i.increaseRetryCount(ctx, item)
	}
	patchable := &nftitem.PatchableNftItem{
		Attributes:        attrs,
		IndexerState:      (*nftitem.IndexerState)(ptr.String(nftitem.IndexerStateFetchingAnimation)),
		IndexerRetryCount: ptr.Int32(0),
	}

	if err := i.tokenUC.PatchNft(ctx, item.ToId(), patchable); err != nil {
		ctx.WithField("err", err).Error("token.PatchNft failed")
		return err
	}

	return nil
}

func (i *NftIndexer) storeAnimationUrl(ctx bCtx.Ctx, item *nftitem.NftItem) error {
	// webresource related failures are not considered errors, only db failures are
	ctx.Info("storeAnimationUrl")
	tokenUri := item.TokenUri
	if len(item.HostedTokenUri) > 0 {
		tokenUri = item.HostedTokenUri
	}
	metadata, err := i.webResourceUC.GetJson(ctx, tokenUri)
	if err != nil {
		ctx.WithField("err", err).Error("webresource.Get failed")
		return i.increaseRetryCount(ctx, item)
	}
	url, err := i.animationUrlParserSelector.GetParser(collection.CollectionId{
		ChainId: item.ChainId,
		Address: item.ContractAddress,
	}).GetAnimationUrl(ctx, metadata, "")
	if err != nil {
		ctx.WithField("err", err).Error("json.Unmarshal failed")
		return i.increaseRetryCount(ctx, item)
	}
	patchable := &nftitem.PatchableNftItem{
		IndexerState:      (*nftitem.IndexerState)(ptr.String(nftitem.IndexerStateDone)),
		IndexerRetryCount: ptr.Int32(0),
	}
	if len(url) > 0 {
		// if has animation_url, fetch and store it, otherwise set state to done
		data, err := i.webResourceUC.Get(ctx, url)
		if err != nil {
			ctx.WithFields(log.Fields{
				"url": url,
				"err": err,
			}).Error("webresource.Get failed")
			return i.increaseRetryCount(ctx, item)
		}

		// detect mime type
		mtype := mimetype.Detect(data)
		mimeType := mtype.String()
		ext := mtype.Extension()

		path, err := i.webResourceUC.Store(
			ctx, item.ChainId, item.ContractAddress, item.TokenId,
			"animation", ext, data, mimeType,
		)
		if err != nil {
			ctx.WithField("err", err).Error("webresource.Store failed")
			return i.increaseRetryCount(ctx, item)
		}
		patchable.AnimationUrl = &url
		patchable.HostedAnimationUrl = &path
		patchable.AnimationUrlContentType = ptr.String(mimeToContentType(mimeType))
		patchable.AnimationUrlMimeType = &mimeType
	}
	if err := i.tokenUC.PatchNft(ctx, item.ToId(), patchable); err != nil {
		ctx.WithField("err", err).Error("token.PatchNft failed")
		return err
	}
	return nil
}

func (i *NftIndexer) increaseRetryCount(ctx bCtx.Ctx, item *nftitem.NftItem) error {
	patchable := &nftitem.PatchableNftItem{
		IndexerRetryCount: ptr.Int32(item.IndexerRetryCount + 1),
	}
	if err := i.tokenUC.PatchNft(ctx, item.ToId(), patchable); err != nil {
		ctx.WithField("err", err).Error("token.PatchNft failed")
		return err
	}
	return nil
}

func mimeToContentType(_typ string) string {
	if _typ == "image/gif" {
		return "gif"
	}
	if strings.HasPrefix(_typ, "video") {
		return "video"
	}
	if strings.HasPrefix(_typ, "audio") {
		return "video"
	}
	return "image"
}
