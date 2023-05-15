package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"

	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/base/ethereum"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/base/nft_indexer"
	pricefomatter "github.com/x-xyz/goapi/base/price_fomatter"
	"github.com/x-xyz/goapi/base/tracker"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/erc1155"
	"github.com/x-xyz/goapi/domain/erc721/contract"
	"github.com/x-xyz/goapi/domain/nftitem"
	mmiddleware "github.com/x-xyz/goapi/middleware"
	"github.com/x-xyz/goapi/service/chain"
	serviceContract "github.com/x-xyz/goapi/service/chain/contract"
	"github.com/x-xyz/goapi/service/chainlink"
	"github.com/x-xyz/goapi/service/coingecko"
	"github.com/x-xyz/goapi/service/query"
	accountRepo "github.com/x-xyz/goapi/stores/account/repository"
	accountUsecase "github.com/x-xyz/goapi/stores/account/usecase"

	cRepo "github.com/x-xyz/goapi/stores/chain/repository"
	cUseCase "github.com/x-xyz/goapi/stores/chain/usecase"
	chainlinkUseCase "github.com/x-xyz/goapi/stores/chainlink/usecase"
	colRepo "github.com/x-xyz/goapi/stores/collection/repository"
	colUseCase "github.com/x-xyz/goapi/stores/collection/usecase"
	order_usecase "github.com/x-xyz/goapi/stores/order/usecase"

	apecoinstakingRepo "github.com/x-xyz/goapi/stores/apecoinstaking/repository"
	apecoinstakingUseCase "github.com/x-xyz/goapi/stores/apecoinstaking/usecase"
	erc1155Repo "github.com/x-xyz/goapi/stores/erc1155/repository"
	erc1155UseCase "github.com/x-xyz/goapi/stores/erc1155/usecase"
	e7UseCase "github.com/x-xyz/goapi/stores/erc721/usecase"
	exchangeUseCase "github.com/x-xyz/goapi/stores/exchange/usecase"
	order_repo "github.com/x-xyz/goapi/stores/order/repository"
	ptRepo "github.com/x-xyz/goapi/stores/paytoken/repository"
	punkUseCase "github.com/x-xyz/goapi/stores/punk/usecase"

	"github.com/x-xyz/goapi/stores/token/repository"
	tokenUseCase "github.com/x-xyz/goapi/stores/token/usecase"
	"github.com/x-xyz/goapi/stores/tracker_state/repository/mongo"
	"github.com/x-xyz/goapi/stores/tracker_state/usecase"
)

func init() {
	viper.SetConfigType("yaml")
	viper.SetConfigFile(`infra/configs/tracker/config.yaml`)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	if viper.GetBool(`debug`) {
		log.Log().Info("Service RUN on DEBUG mode")
	}

	// overwrite active network in the config if the environment has been set
	viper.BindEnv("ACTIVENETWORK")
}

func main() {
	ctx, cancel := bCtx.WithCancel(bCtx.Background())

	// start server to pass cloud run health check
	startEchoServer()

	ctxTimeout := viper.GetDuration("context.timeout")
	checkNewContractInterval := viper.GetDuration("tracker.checkNewContractInterval")
	followDistance := viper.GetUint64("tracker.followDistance")
	activeNetwork := viper.GetString("activeNetwork")
	networkInfo := viper.Sub(fmt.Sprintf("networks.%s", activeNetwork))
	chainId := networkInfo.GetInt64("chainId")
	blockTime := networkInfo.GetDuration("blockTime")
	wsUrl := networkInfo.GetString("wsUrl")
	rpcUrl := networkInfo.GetString("rpcUrl")
	archiveRpcUrl := networkInfo.GetString("archiveRpcUrl")
	indexerInfo := viper.Sub("indexer")
	indexerRetryLimit := indexerInfo.GetInt("retryLimit")
	indexerBatch := indexerInfo.GetInt("batch")
	indexerWorkers := indexerInfo.GetInt("workers")
	indexerInterval := indexerInfo.GetDuration("interval")
	metatdataInterval := indexerInfo.GetDuration("metadataInterval")

	contractInfo := viper.Sub(fmt.Sprintf("contract.%s", activeNetwork))
	exchangeContract := contractInfo.GetString("exchange")
	manifoldContract := contractInfo.GetString("manifold.royaltyRegistry")
	apecoinStakingContract := "0x5954ab967bc958940b7eb73ee84797dc8a2afbb9"
	royaltyEngineContrct := contractInfo.GetString("royaltyEngine")
	priceUpdaterInterval := viper.GetDuration("priceUpdater.interval")

	ctx.WithFields(log.Fields{
		"network":          activeNetwork,
		"chainId":          chainId,
		"blockTime":        blockTime,
		"wsUrl":            wsUrl,
		"rpcUrl":           rpcUrl,
		"acrhiveRpcUrl":    archiveRpcUrl,
		"exchangeContract": exchangeContract,
	}).Info("config")

	ctx.Info("init mongo")
	q := initMongo()
	ctx.Info("connecting eth clients")
	wsClient, rpcClient, archiveEthClient := initEthClient(ctx, wsUrl, rpcUrl, archiveRpcUrl)
	_clientProvider := newClientProvider(ctx, 15, wsUrl)
	throttledClient := ethereum.NewTrottledClient(rpcClient, 100)
	errCh := make(chan error, 10)
	chainService, err := chain.NewClient(ctx, &chain.ClientCfg{
		RpcUrls: map[int32]string{
			int32(chainId): rpcUrl,
		},
		ArchiveRpcUrls: map[int32]string{
			int32(chainId): archiveRpcUrl,
		},
	})
	if err != nil {
		ctx.WithField("err", err).Panic("chainService init failed")
	}
	chainlinkService := chainlink.New(chainService)
	coinGecko := coingecko.NewClient(&coingecko.ClientCfg{
		HttpClient: http.Client{},
		Timeout:    10 * time.Second,
	})

	var trackers []*tracker.EventTracker
	var needUpdateIndexerStates = []nftitem.IndexerState{
		nftitem.IndexerStateHasTokenURI,
		nftitem.IndexerStateHasTokenURIRefreshing,
		nftitem.IndexerStateHasImageURL,
		nftitem.IndexerStateHasHostedImage,
		nftitem.IndexerStateParsingAttributes,
		nftitem.IndexerStateFetchingAnimation,
		nftitem.IndexerStateDone,
	}

	// repos
	nftitemRepo := repository.NewNftItem(q, nil)
	paytokenRepo := ptRepo.NewPayTokenRepo(q)
	trackerStateRepo := mongo.NewTrackerStateMongoRepo(q)
	blockRepo := cRepo.NewBlockRepo(q)
	erc721Repo := colRepo.NewErc721Contract(q)
	erc1155ContractRepo := erc1155Repo.NewContractRepo(q)
	erc1155HoldingRepo := erc1155Repo.NewHoldingRepo(q)
	collectionRepo := colRepo.NewCollection(q)
	activityHistoryRepo := accountRepo.NewActivityHistoryRepo(q)
	tradingVolumeRepo := colRepo.NewTradingVolumeRepo(q)
	floorPriceHistoryRepo := colRepo.NewFloorPriceHistoryRepo(q)
	folderRepo := accountRepo.NewFolderRepo(q)
	folderNftRelationshipRepo := accountRepo.NewFolderNftRelationshipRepo(q)
	orderRepo := order_repo.NewOrderRepo(q)
	orderItemRepo := order_repo.NewOrderItemRepo(q)
	orderNonceRepo := accountRepo.NewOrderNonceRepo(q)
	apecoinStakingRepo := apecoinstakingRepo.New(q)

	// usecases
	tokenUC := tokenUseCase.New(&tokenUseCase.TokenUseCaseCfg{
		NftitemRepo:    nftitemRepo,
		CollectionRepo: collectionRepo,
		OrderItemRepo:  orderItemRepo,
	})
	chainlinkUC := chainlinkUseCase.New(chainlinkService, paytokenRepo)
	tradingVolumeUC := colUseCase.NewTradingVolumeUseCase(tradingVolumeRepo, chainlinkUC)
	blockUseCase := cUseCase.NewBlockUseCase(blockRepo)
	colUC := colUseCase.NewCollection(&colUseCase.CollectionUseCaseCfg{
		CollectionRepo:        collectionRepo,
		FloorPriceHistoryRepo: floorPriceHistoryRepo,
		OrderItemRepo:         orderItemRepo,
	})
	priceFormatter := pricefomatter.NewPriceFormatter(&pricefomatter.PriceFormatterCfg{
		Paytoken:  paytokenRepo,
		Chainlink: chainlinkUC,
		CoinGecko: coinGecko,
	})
	order := order_usecase.New(&order_usecase.OrderUseCaseCfg{
		ExchangeCfgs:        nil,
		OrderRepo:           orderRepo,
		OrderItemRepo:       orderItemRepo,
		NftitemRepo:         nftitemRepo,
		Erc1155HoldingRepo:  erc1155HoldingRepo,
		AccountRepo:         nil,
		PaytokenRepo:        paytokenRepo,
		PriceFormatter:      priceFormatter,
		OrderNonceUC:        nil,
		Erc1271:             nil,
		ActivityHistoryRepo: activityHistoryRepo,
	})
	orderNonceUC := accountUsecase.NewOrderNonceUseCase(orderNonceRepo)
	exchangeUC := exchangeUseCase.NewExchangeUseCase(&exchangeUseCase.ExchangeUseCaseCfg{
		OrderUseCase:      order,
		OrderNonceUseCase: orderNonceUC,
		Nftitem:           nftitemRepo,
		Token:             tokenUC,
		ActivityHistory:   activityHistoryRepo,
		Collection:        colUC,
		TradingVolume:     tradingVolumeUC,
		PriceFormatter:    priceFormatter,
	})
	tsUseCase := usecase.NewTrackerStateUseCase(trackerStateRepo, ctxTimeout)
	folderUsecase := accountUsecase.NewFolderUsecase(
		folderRepo,
		folderNftRelationshipRepo,
		nftitemRepo,
		collectionRepo,
		floorPriceHistoryRepo,
		coinGecko,
		erc1155HoldingRepo,
	)
	erc721UseCase := e7UseCase.NewErc721UseCase(erc721Repo)
	erc721EventUseCase := e7UseCase.NewErc721EventUseCase(&e7UseCase.Erc721EventUseCaseCfg{
		Nftitem:                   nftitemRepo,
		Erc721:                    erc721Repo,
		Token:                     tokenUC,
		CollectionRepo:            collectionRepo,
		FolderRepo:                folderRepo,
		FolderNftRelationshipRepo: folderNftRelationshipRepo,
		ActivityHistoryRepo:       activityHistoryRepo,
		FolderUsecase:             folderUsecase,
		OrderUseCase:              order,
	})
	erc1155UC := erc1155UseCase.NewErc1155UseCase(erc1155ContractRepo)
	erc1155EventUseCase := erc1155UseCase.NewErc1155EventUseCase(&erc1155UseCase.Erc1155EventUseCaseCfg{
		Nftitem:             nftitemRepo,
		Erc1155:             erc1155ContractRepo,
		Holding:             erc1155HoldingRepo,
		CollectionRepo:      collectionRepo,
		ActivityHistoryRepo: activityHistoryRepo,
		Token:               tokenUC,
		FolderUsecase:       folderUsecase,
		OrderUseCase:        order,
	})
	punkEventUseCase := punkUseCase.NewPunkEventUseCase(&punkUseCase.PunkEventUseCaseCfg{
		Nftitem:        nftitemRepo,
		CollectionRepo: collectionRepo,
	})
	apecoinStakingUC := apecoinstakingUseCase.New(apecoinStakingRepo)

	// handlers
	exchangeHandler := tracker.NewExchangeEventHandler(&tracker.ExchangeEventHandlerCfg{
		ChainId:         chainId,
		ExchangeUseCase: exchangeUC,
	})
	erc721Handler := tracker.NewErc721EventHandler(&tracker.Erc721EventHandlerCfg{
		ChainId:            chainId,
		Erc721EventUseCase: erc721EventUseCase,
	})
	erc1155Handler := tracker.NewErc1155EventHandler(&tracker.Erc1155EventHandlerCfg{
		ChainId:             chainId,
		Erc1155EventUseCase: erc1155EventUseCase,
	})
	punkHandler := tracker.NewPunkEventHandler(&tracker.PunkEventHandlerCfg{
		ChainId:          chainId,
		PunkEventUseCase: punkEventUseCase,
		RpcClient:        throttledClient,
	})
	manifoldEventHandler := tracker.NewManifoldEventHandler(&tracker.ManifoldEventHandlerCfg{
		ChainId:        domain.ChainId(chainId),
		CollectionRepo: collectionRepo,
		NftItemRepo:    nftitemRepo,
		ChainService:   chainService,
		RoyaltyEngine:  common.HexToAddress(royaltyEngineContrct),
	})
	apecoinStakingEventHandler := tracker.NewApecoinStakingEventHandler(&tracker.ApecoinStakingEventHandlerCfg{
		ChainId:                chainId,
		ApecoinStakingUC:       apecoinStakingUC,
		ApecoinStakingContract: serviceContract.NewApecoinStaking(chainService, int32(chainId), common.HexToAddress(apecoinStakingContract)),
		TokenUC:                tokenUC,
	})

	currentBlockGetter := tracker.NewCurrentBlockGetter(&tracker.CurrentBlockGetterCfg{
		Client: wsClient,
		ErrCh:  errCh,
	})
	// trackers
	exchangeTracker, err := tracker.NewEventTracker(&tracker.EventTrackerCfg{
		ChainId:             chainId,
		BlockTime:           blockTime,
		CurrentBlockGetter:  currentBlockGetter,
		Mongo:               q,
		WsClient:            _clientProvider.consume(ctx),
		RpcClient:           throttledClient,
		ClientWithArchive:   archiveEthClient,
		TrackerStateUseCase: tsUseCase,
		TrackerTag:          "exchange",
		ShouldDecodeSender:  false,
		FollowDistance:      followDistance,
		BlockUseCase:        blockUseCase,
		ContractAddress:     common.HexToAddress(exchangeContract),
		EventHandl:          exchangeHandler,
		ErrorCh:             errCh,
	})
	if err != nil {
		ctx.WithField("err", err).Panic("new exchange tracker failed")
	}
	trackers = append(trackers, exchangeTracker)

	manifoldTracker, err := tracker.NewEventTracker(&tracker.EventTrackerCfg{
		ChainId:             chainId,
		BlockTime:           blockTime,
		CurrentBlockGetter:  currentBlockGetter,
		Mongo:               q,
		WsClient:            _clientProvider.consume(ctx),
		RpcClient:           throttledClient,
		ClientWithArchive:   archiveEthClient,
		TrackerStateUseCase: tsUseCase,
		TrackerTag:          domain.DefaultTag,
		ShouldDecodeSender:  false,
		FollowDistance:      followDistance,
		BlockUseCase:        blockUseCase,
		ContractAddress:     common.HexToAddress(manifoldContract),
		EventHandl:          manifoldEventHandler,
		ErrorCh:             errCh,
		SkipMissingBlock:    false,
	})
	if err != nil {
		ctx.WithField("err", err).Panic("new manifold tracker failed")
	}
	trackers = append(trackers, manifoldTracker)

	if chainId == 1 {
		apecoinStakingTracker, err := tracker.NewEventTracker(&tracker.EventTrackerCfg{
			ChainId:             chainId,
			BlockTime:           blockTime,
			CurrentBlockGetter:  currentBlockGetter,
			Mongo:               q,
			WsClient:            _clientProvider.consume(ctx),
			RpcClient:           throttledClient,
			ClientWithArchive:   archiveEthClient,
			TrackerStateUseCase: tsUseCase,
			TrackerTag:          domain.DefaultTag,
			ShouldDecodeSender:  false,
			FollowDistance:      followDistance,
			BlockUseCase:        blockUseCase,
			ContractAddress:     common.HexToAddress(apecoinStakingContract),
			EventHandl:          apecoinStakingEventHandler,
			ErrorCh:             errCh,
			SkipMissingBlock:    false,
		})
		if err != nil {
			ctx.WithField("err", err).Panic("new apecoin staking tracker failed")
		}
		trackers = append(trackers, apecoinStakingTracker)
	}

	// token trackers
	trackingTokens := make(map[domain.Address]struct{})
	tokens, err := erc721UseCase.FindAll(ctx, contract.WithChainId(domain.ChainId(chainId)), contract.WithIsAppropriate(true))
	if err != nil {
		ctx.WithField("err", err).Panic("erc721Repo.FindAll failed")
	}
	ctx.Info(fmt.Sprintf("%d erc721 contracts", len(tokens)))
	for _, t := range tokens {
		ctx.WithField("contract", t.Address).Info("tracking erc721 contract")
		trackingTokens[t.Address] = struct{}{}
		cfg := &tracker.EventTrackerCfg{
			ChainId:             chainId,
			BlockTime:           blockTime,
			CurrentBlockGetter:  currentBlockGetter,
			Mongo:               q,
			WsClient:            _clientProvider.consume(ctx),
			RpcClient:           throttledClient,
			ClientWithArchive:   archiveEthClient,
			TrackerStateUseCase: tsUseCase,
			TrackerTag:          domain.DefaultTag,
			ShouldDecodeSender:  false,
			FollowDistance:      followDistance,
			BlockUseCase:        blockUseCase,
			ContractAddress:     common.HexToAddress(t.Address.ToLowerStr()),
			EventHandl:          erc721Handler,
			ErrorCh:             errCh,
		}
		if t.Address.Equals(domain.PunkAddress) {
			cfg.EventHandl = punkHandler
		}
		tracker, err := tracker.NewEventTracker(cfg)
		if err != nil {
			ctx.WithField("err", err).Panic("new erc721 tracker failed")
		}
		trackers = append(trackers, tracker)
	}

	erc1155Tokens, err := erc1155UC.FindAll(ctx, erc1155.WithChainId(domain.ChainId(chainId)), erc1155.WithIsAppropriate(true))
	if err != nil {
		ctx.WithField("err", err).Panic("erc1155.FindAll failed")
	}
	ctx.Info(fmt.Sprintf("%d erc1155 contracts", len(erc1155Tokens)))
	for _, t := range erc1155Tokens {
		ctx.WithField("erc1155 contract", t.Address).Info("tracking erc1155 contract")
		trackingTokens[t.Address] = struct{}{}
		tracker, err := tracker.NewEventTracker(&tracker.EventTrackerCfg{
			ChainId:             chainId,
			BlockTime:           blockTime,
			CurrentBlockGetter:  currentBlockGetter,
			Mongo:               q,
			WsClient:            _clientProvider.consume(ctx),
			RpcClient:           throttledClient,
			ClientWithArchive:   archiveEthClient,
			TrackerStateUseCase: tsUseCase,
			TrackerTag:          domain.DefaultTag,
			ShouldDecodeSender:  false,
			FollowDistance:      followDistance,
			BlockUseCase:        blockUseCase,
			ContractAddress:     common.HexToAddress(t.Address.ToLowerStr()),
			EventHandl:          erc1155Handler,
			ErrorCh:             errCh,
		})
		if err != nil {
			ctx.WithField("err", err).Panic("new erc1155 tracker failed")
		}
		trackers = append(trackers, tracker)
	}

	nftTokenURIIndexer := nft_indexer.NewNftTokenURIIndexer(&nft_indexer.NftTokenURIIndexerCfg{
		TokenUC:     tokenUC,
		ChainId:     domain.ChainId(chainId),
		EthClient:   throttledClient,
		TargetState: nftitem.IndexerStateNew,
		RetryLimit:  indexerRetryLimit,
		Batch:       indexerBatch,
		Workers:     indexerWorkers,
		Interval:    indexerInterval,
		ErrorCh:     errCh,
	})
	nftTokenURIRefreshingIndexer := nft_indexer.NewNftTokenURIIndexer(&nft_indexer.NftTokenURIIndexerCfg{
		TokenUC:     tokenUC,
		ChainId:     domain.ChainId(chainId),
		EthClient:   throttledClient,
		TargetState: nftitem.IndexerStateNewRefreshing,
		RetryLimit:  indexerRetryLimit,
		Batch:       indexerBatch,
		Workers:     indexerWorkers,
		Interval:    indexerInterval,
		ErrorCh:     errCh,
	})
	priceUpdater := tracker.NewPriceUpdater(&tracker.PriceUpdaterCfg{
		ChainId:        domain.ChainId(chainId),
		Collection:     colUC,
		Token:          tokenUC,
		Order:          order,
		Interval:       priceUpdaterInterval,
		PriceFormatter: priceFormatter,
		ErrorCh:        errCh,
	})
	metadataRefreshingIndexer := nft_indexer.NewMetadataUpdater(&nft_indexer.MetadataUpdaterCfg{
		TokenUC:      tokenUC,
		ChainId:      domain.ChainId(chainId),
		TargetStates: needUpdateIndexerStates,
		Collections:  domain.MetadataRefreshCollectionAddresses,
		RetryLimit:   indexerRetryLimit,
		Batch:        indexerBatch,
		Workers:      indexerWorkers,
		Interval:     metatdataInterval,
		ErrorCh:      errCh,
	})

	ctx.Info("starting workers")
	err = currentBlockGetter.(*tracker.CurrentBlockGetter).Start(ctx)
	if err != nil {
		ctx.WithField("err", err).Panic("currentBlockGetter.Start failed")
	}
	for _, t := range trackers {
		t.Start(ctx)
	}
	nftTokenURIIndexer.Start(ctx)
	nftTokenURIRefreshingIndexer.Start(ctx)
	priceUpdater.Start(ctx)
	metadataRefreshingIndexer.Start(ctx)

	ticker := time.NewTicker(checkNewContractInterval)
	defer ticker.Stop()
FOR:
	for {
		select {
		case err := <-errCh:
			ctx.WithField("err", err).Error("tracker error")
			break FOR
		case <-ticker.C:
			ctx.Info("checking for new contracts")
			tokens, err := erc721UseCase.FindAll(ctx, contract.WithChainId(domain.ChainId(chainId)), contract.WithIsAppropriate(true))
			if err != nil {
				ctx.WithField("err", err).Panic("erc721Repo.FindAll failed")
				break FOR
			}
			for _, t := range tokens {
				if _, ok := trackingTokens[t.Address]; !ok {
					ctx.WithField("contract", t.Address).Info("tracking contract")
					trackingTokens[t.Address] = struct{}{}

					// ignore punk here, restart tracker to index punk for the first time
					if t.Address.Equals(domain.PunkAddress) {
						continue
					}

					tracker, err := tracker.NewEventTracker(&tracker.EventTrackerCfg{
						ChainId:             chainId,
						BlockTime:           blockTime,
						CurrentBlockGetter:  currentBlockGetter,
						Mongo:               q,
						WsClient:            _clientProvider.consume(ctx),
						RpcClient:           throttledClient,
						ClientWithArchive:   archiveEthClient,
						TrackerStateUseCase: tsUseCase,
						TrackerTag:          domain.DefaultTag,
						ShouldDecodeSender:  false,
						FollowDistance:      followDistance,
						BlockUseCase:        blockUseCase,
						ContractAddress:     common.HexToAddress(t.Address.ToLowerStr()),
						EventHandl:          erc721Handler,
						ErrorCh:             errCh,
					})
					if err != nil {
						ctx.WithField("err", err).Panic("new erc721 tracker failed")
					}
					tracker.Start(ctx)
					trackers = append(trackers, tracker)
				}
			}

			ctx.Info("checking for new erc1155 contracts")
			erc1155Tokens, err := erc1155UC.FindAll(ctx, erc1155.WithChainId(domain.ChainId(chainId)), erc1155.WithIsAppropriate(true))
			if err != nil {
				ctx.WithField("err", err).Panic("erc1155.FindAll failed")
				break FOR
			}
			for _, t := range erc1155Tokens {
				if _, ok := trackingTokens[t.Address]; !ok {
					ctx.WithField("erc1155 contract", t.Address).Info("tracking erc1155 contract")
					trackingTokens[t.Address] = struct{}{}
					tracker, err := tracker.NewEventTracker(&tracker.EventTrackerCfg{
						ChainId:             chainId,
						BlockTime:           blockTime,
						CurrentBlockGetter:  currentBlockGetter,
						Mongo:               q,
						WsClient:            _clientProvider.consume(ctx),
						RpcClient:           throttledClient,
						ClientWithArchive:   archiveEthClient,
						TrackerStateUseCase: tsUseCase,
						TrackerTag:          domain.DefaultTag,
						ShouldDecodeSender:  false,
						FollowDistance:      followDistance,
						BlockUseCase:        blockUseCase,
						ContractAddress:     common.HexToAddress(t.Address.ToLowerStr()),
						EventHandl:          erc1155Handler,
						ErrorCh:             errCh,
					})
					if err != nil {
						ctx.WithField("err", err).Panic("new erc1155 tracker failed")
					}
					tracker.Start(ctx)
					trackers = append(trackers, tracker)
				}
			}
		}
	}

	go func() {
		for range errCh {
		}
	}()
	cancel()

	priceUpdater.Wait()
	nftTokenURIIndexer.Wait()
	nftTokenURIRefreshingIndexer.Wait()
	for _, t := range trackers {
		t.Wait()
	}
	currentBlockGetter.(*tracker.CurrentBlockGetter).Wait()
	metadataRefreshingIndexer.Wait()
}

func startEchoServer() {
	context := bCtx.Background()

	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{}))
	e.Use(middleware.RequestID())
	middL := mmiddleware.InitMiddleware()
	e.Use(middL.ResponseLogger())
	e.Use(middL.AddContext())

	address := viper.GetString("server.address")
	context.WithField("address", address).Info("starting server")
	go func() {
		if err := e.Start(address); err != nil && err != http.ErrServerClosed {
			context.Error("shutting down the server")
		}
	}()
}

func initMongo() query.Mongo {
	uri := viper.GetString("mongo.uri")
	authDBName := viper.GetString("mongo.authDBName")
	dbName := viper.GetString("mongo.dbName")
	enableSSL := viper.GetBool("mongo.enableSSL")
	checkIndex := viper.GetBool("mongo.checkIndex")
	mongoClient := mongoclient.MustConnectMongoClient(uri, authDBName, dbName, enableSSL, true, 2)
	return query.New(mongoClient, checkIndex)
}

func initEthClient(ctx bCtx.Ctx, rpcUrl, secondaryUrl, archiveRpcUrl string) (*ethclient.Client, *ethclient.Client, *ethclient.Client) {
	client, err := ethclient.DialContext(ctx, rpcUrl)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
			"url": rpcUrl,
		}).Panic("failed to connect rpc")
	}

	secondaryClient, err := ethclient.DialContext(ctx, secondaryUrl)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
			"url": secondaryUrl,
		}).Panic("failed to connect secondary rpc")
	}

	archiveClient, err := ethclient.DialContext(ctx, archiveRpcUrl)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
			"url": archiveRpcUrl,
		}).Panic("failed to connect archive rpc")
	}

	return client, secondaryClient, archiveClient
}
