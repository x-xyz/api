package nft_indexer

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/x-xyz/goapi/base/abi"
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/base/metrics"
	"github.com/x-xyz/goapi/base/ptr"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/nftitem"
	"github.com/x-xyz/goapi/domain/token"
)

type NftTokenURIIndexerCfg struct {
	TokenUC     token.Usecase
	ChainId     domain.ChainId
	EthClient   domain.EthClientRepo
	TargetState nftitem.IndexerState
	RetryLimit  int
	Batch       int
	Workers     int
	Interval    time.Duration
	ErrorCh     chan<- error
}

type NftTokenURIIndexer struct {
	tokenUC     token.Usecase
	chainId     domain.ChainId
	ethClient   domain.EthClientRepo
	targetState nftitem.IndexerState
	retryLimit  int
	batch       int
	workers     int
	interval    time.Duration
	taskCh      chan *nftitem.NftItem
	errorCh     chan<- error
	stoppedCh   chan interface{}
}

func NewNftTokenURIIndexer(cfg *NftTokenURIIndexerCfg) *NftTokenURIIndexer {
	metOnce.Do(func() {
		met = metrics.New("indexer")
	})
	return &NftTokenURIIndexer{
		tokenUC:     cfg.TokenUC,
		chainId:     cfg.ChainId,
		ethClient:   cfg.EthClient,
		targetState: cfg.TargetState,
		retryLimit:  cfg.RetryLimit,
		batch:       cfg.Batch,
		workers:     cfg.Workers,
		interval:    cfg.Interval,
		taskCh:      make(chan *nftitem.NftItem, cfg.Batch),
		errorCh:     cfg.ErrorCh,
		stoppedCh:   make(chan interface{}),
	}
}

func (i *NftTokenURIIndexer) Start(ctx bCtx.Ctx) {
	go i.loop(ctx)
}

func (i *NftTokenURIIndexer) Wait() {
	<-i.stoppedCh
}

func (i *NftTokenURIIndexer) loop(ctx bCtx.Ctx) {
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
				token.WithChainId(i.chainId),
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

func (i *NftTokenURIIndexer) processNft(ctx bCtx.Ctx, item *nftitem.NftItem) error {
	ctx = bCtx.WithValues(ctx, map[string]interface{}{
		"chainId":  item.ChainId,
		"contract": item.ContractAddress,
		"tokenId":  item.TokenId,
		"tokenURI": item.TokenUri,
		"imageUrl": item.ImageUrl,
	})
	switch item.IndexerState {
	case nftitem.IndexerStateNew:
		return i.getTokenURI(ctx, item)
	case nftitem.IndexerStateNewRefreshing:
		return i.getTokenURI(ctx, item)
	default:
		ctx.WithField("state", item.IndexerState).Warn("unknown state")
		return nil
	}
}

func (i *NftTokenURIIndexer) getTokenURI(ctx bCtx.Ctx, item *nftitem.NftItem) error {
	// ethclient related failures are not considered errors, only db failures are
	// set tokenURI
	ctx.Info("getTokenURI")
	uri, err := i.getTokenURIFromContract(ctx, item.TokenType, item.ContractAddress, item.TokenId)
	if err != nil {
		ctx.WithField("err", err).Error("getTokenURIFromContract failed")
		return i.increaseRetryCount(ctx, item)
	}

	var patchable = &nftitem.PatchableNftItem{}

	if item.IndexerState == nftitem.IndexerStateNew {
		patchable = &nftitem.PatchableNftItem{
			TokenUri:          ptr.String(uri),
			IndexerState:      (*nftitem.IndexerState)(ptr.String(nftitem.IndexerStateHasTokenURI)),
			IndexerRetryCount: ptr.Int32(0),
		}
	}
	if item.IndexerState == nftitem.IndexerStateNewRefreshing {
		patchable = &nftitem.PatchableNftItem{
			TokenUri:          ptr.String(uri),
			IndexerState:      (*nftitem.IndexerState)(ptr.String(nftitem.IndexerStateHasTokenURIRefreshing)),
			IndexerRetryCount: ptr.Int32(0),
		}
	}

	item.TokenUri = uri
	if err := i.tokenUC.PatchNft(ctx, item.ToId(), patchable); err != nil {
		ctx.WithField("err", err).Error("token.PatchNft failed")
	}
	return nil
}

func (i *NftTokenURIIndexer) getTokenURIFromContract(ctx bCtx.Ctx, tokenType domain.TokenType, contractAddress domain.Address, tokenId domain.TokenId) (string, error) {
	switch contractAddress.ToLower() {
	// ens
	case "0x57f1887a8bf19b14fc0df6fd9b2acc9af147ea85":
		return i.getURIFromENS(ctx, contractAddress, tokenId)

	// town star
	case "0xc36cf0cfcb5d905b8b513860db0cfe63f6cf9f5c":
		return i.getURIFromTownStar(ctx, contractAddress, tokenId)

	// decentraland
	case "0xf87e31492faf9a91b02ee0deaad50d51d56d5d4d":
		return i.getURIFromDecentraland(ctx, contractAddress, tokenId)
	default:
	}

	switch tokenType {
	case domain.TokenType721:
		return i.getURIFrom721(ctx, contractAddress, tokenId)
	case domain.TokenType1155:
		return i.getURIFrom1155(ctx, contractAddress, tokenId)
	case domain.TokenTypePunk:
		return i.getURIFromPunkData(ctx, tokenId)
	}
	return "", errors.New("unsupported token type")
}

func (i *NftTokenURIIndexer) getURIFromTownStar(ctx bCtx.Ctx, contracAddress domain.Address, tokenId domain.TokenId) (string, error) {
	townStarEndpoint := fmt.Sprintf("https://tokens.gala.games/metadata/0xc36cf0cfcb5d905b8b513860db0cfe63f6cf9f5c/%s", tokenId)
	return townStarEndpoint, nil
}

func (i *NftTokenURIIndexer) getURIFromENS(ctx bCtx.Ctx, contractAddress domain.Address, tokenId domain.TokenId) (string, error) {
	ensEndpoint := fmt.Sprintf("https://metadata.ens.domains/mainnet/0x57f1887a8bf19b14fc0df6fd9b2acc9af147ea85/%s", tokenId)
	return ensEndpoint, nil
}

func (i *NftTokenURIIndexer) getURIFromDecentraland(ctx bCtx.Ctx, contractAddress domain.Address, tokenId domain.TokenId) (string, error) {
	decentralandEndpoint := fmt.Sprintf("https://api.decentraland.org/v2/contracts/0xf87e31492faf9a91b02ee0deaad50d51d56d5d4d/tokens/%s", tokenId)
	return decentralandEndpoint, nil
}

func (i *NftTokenURIIndexer) getURIFrom721(ctx bCtx.Ctx, contractAddress domain.Address, tokenId domain.TokenId) (string, error) {
	method := "tokenURI"
	addr := common.HexToAddress(contractAddress.ToLowerStr())
	_tokenId, ok := new(big.Int).SetString(tokenId.String(), 10)
	if !ok {
		err := errors.New("big.Int.SetString failed")
		ctx.WithField("tokenId", tokenId).Error(err.Error())
		return "", err
	}
	data, err := abi.ERC721TokenABI.Pack(method, _tokenId)
	if err != nil {
		ctx.WithField("err", err).Error("ERC721TokenABI.Pack failed")
		return "", err
	}
	msg := ethereum.CallMsg{
		To:   &addr,
		Data: data,
	}
	res, err := i.ethClient.CallContract(ctx, msg, nil)
	if err != nil {
		ctx.WithField("err", err).Error("ethclient.CallContract failed")
		return "", err
	}
	unpacked, err := abi.ERC721TokenABI.Unpack(method, res)
	if err != nil {
		ctx.WithField("err", err).Error("ERC721TokenABI.Unpack failed")
		return "", err
	}
	return unpacked[0].(string), nil
}

func (i *NftTokenURIIndexer) getURIFrom1155(ctx bCtx.Ctx, contractAddress domain.Address, tokenId domain.TokenId) (string, error) {
	method := "uri"
	addr := common.HexToAddress(contractAddress.ToLowerStr())
	_tokenId, ok := new(big.Int).SetString(tokenId.String(), 10)
	if !ok {
		err := errors.New("big.Int.SetString failed")
		ctx.WithField("tokenId", tokenId).Error(err.Error())
		return "", err
	}
	data, err := abi.ERC1155TokenABI.Pack(method, _tokenId)
	if err != nil {
		ctx.WithField("err", err).Error("ERC1155TokenABI.Pack failed")
		return "", err
	}
	msg := ethereum.CallMsg{
		To:   &addr,
		Data: data,
	}
	res, err := i.ethClient.CallContract(ctx, msg, nil)
	if err != nil {
		ctx.WithField("err", err).Error("ethclient.CallContract failed")
		return "", err
	}
	unpacked, err := abi.ERC1155TokenABI.Unpack(method, res)
	if err != nil {
		ctx.WithField("err", err).Error("ERC1155TokenABI.Unpack failed")
		return "", err
	}

	// replace {id} with leading zero padded 64 character hex string
	// see: https://eips.ethereum.org/EIPS/eip-1155 Metadata part
	uri := unpacked[0].(string)
	id, err := tokenId.ToHexString()
	if err != nil {
		return "", err
	}
	uri = strings.ReplaceAll(uri, "{id}", id)
	return uri, nil
}

func (i *NftTokenURIIndexer) getURIFromPunkData(ctx bCtx.Ctx, tokenId domain.TokenId) (string, error) {
	index, ok := new(big.Int).SetString(tokenId.String(), 10)
	if !ok {
		err := errors.New("big.Int.SetString failed")
		ctx.WithField("tokenId", tokenId).Error(err.Error())
		return "", err
	}

	image, err := i.getImageSvgFromPunkData(ctx, uint16(index.Uint64()))
	if err != nil {
		ctx.WithField("tokenId", tokenId).Error(err.Error())
		return "", err
	}

	attributes, err := i.getAttributeFromPunkData(ctx, uint16(index.Uint64()))
	if err != nil {
		ctx.WithField("tokenId", tokenId).Error(err.Error())
		return "", err
	}

	data := struct {
		Image      string             `json:"image"`
		Attributes nftitem.Attributes `json:"attributes"`
	}{
		Image:      image,
		Attributes: attributes,
	}

	s, err := json.Marshal(data)
	if err != nil {
		ctx.WithField("tokenId", tokenId).Error(err.Error())
		return "", err
	}

	return fmt.Sprintf("data:application/json;base64,%s", base64.StdEncoding.EncodeToString(s)), nil
}

// getAttributeFromPunkData
// we also transform the attributes from punk data contract in this function
// ex:
// attributes: Female 2, Earring, Blonde Bob, Green Eye Shadow
// => [
//   {"trait_type": "Female 2", "value": "Female 2"},
//   {"trait_type": "Earring", "value": "Earring"},
//   {"trait_type": "Blonde Bob", "value": "Blonde Bob"},
//   {"trait_type": "Green Eye Shadow", "value": "Green Eye Shadow"},
// ]
func (i *NftTokenURIIndexer) getAttributeFromPunkData(ctx bCtx.Ctx, index uint16) (nftitem.Attributes, error) {
	method := "punkAttributes"
	addr := common.HexToAddress(domain.PunkDataAddress.ToLowerStr())
	data, err := abi.PunkDataABI.Pack(method, index)
	if err != nil {
		ctx.WithField("err", err).Error("punkAttributes.Pack failed")
		return nil, err
	}
	msg := ethereum.CallMsg{
		To:   &addr,
		Data: data,
	}
	res, err := i.ethClient.CallContract(ctx, msg, nil)
	if err != nil {
		ctx.WithField("err", err).Error("ethclient.CallContract failed")
		return nil, err
	}
	unpacked, err := abi.PunkDataABI.Unpack(method, res)
	if err != nil {
		ctx.WithField("err", err).Error("punkAttributes.Unpack failed")
		return nil, err
	}

	attrs := unpacked[0].(string)
	var attributes nftitem.Attributes
	for _, attr := range strings.Split(attrs, ",") {
		name := strings.TrimSpace(attr)
		attributes = append(attributes, nftitem.Attribute{
			TraitType: name,
			Value:     name,
		})
	}
	return attributes, nil
}

func (i *NftTokenURIIndexer) getImageSvgFromPunkData(ctx bCtx.Ctx, index uint16) (string, error) {
	method := "punkImageSvg"
	addr := common.HexToAddress(domain.PunkDataAddress.ToLowerStr())
	data, err := abi.PunkDataABI.Pack(method, index)
	if err != nil {
		ctx.WithField("err", err).Error("punkImageSvg.Pack failed")
		return "", err
	}
	msg := ethereum.CallMsg{
		To:   &addr,
		Data: data,
	}
	res, err := i.ethClient.CallContract(ctx, msg, nil)
	if err != nil {
		ctx.WithField("err", err).Error("ethclient.CallContract failed")
		return "", err
	}
	unpacked, err := abi.PunkDataABI.Unpack(method, res)
	if err != nil {
		ctx.WithField("err", err).Error("punkIamgeSvg.Unpack failed")
		return "", err
	}

	return unpacked[0].(string), nil
}

func (i *NftTokenURIIndexer) increaseRetryCount(ctx bCtx.Ctx, item *nftitem.NftItem) error {
	patchable := &nftitem.PatchableNftItem{
		IndexerRetryCount: ptr.Int32(item.IndexerRetryCount + 1),
	}
	if err := i.tokenUC.PatchNft(ctx, item.ToId(), patchable); err != nil {
		ctx.WithField("err", err).Error("token.PatchNft failed")
		return err
	}
	return nil
}
