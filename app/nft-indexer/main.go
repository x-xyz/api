package main

import (
	"fmt"
	"net/http"

	"cloud.google.com/go/storage"
	ipfsapi "github.com/ipfs/go-ipfs-api"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"
	"github.com/x-xyz/goapi/base/backoff"
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/base/nft_indexer"
	"github.com/x-xyz/goapi/base/nft_indexer/animation_url_parser"
	"github.com/x-xyz/goapi/base/nft_indexer/metadata_parser"
	"github.com/x-xyz/goapi/domain/nftitem"
	mmiddleware "github.com/x-xyz/goapi/middleware"
	"github.com/x-xyz/goapi/service/chain"
	chainlink_service "github.com/x-xyz/goapi/service/chainlink"
	"github.com/x-xyz/goapi/service/opensea"
	"github.com/x-xyz/goapi/service/query"
	account_repository "github.com/x-xyz/goapi/stores/account/repository"
	account_usecase "github.com/x-xyz/goapi/stores/account/usecase"
	apecoinstakingRepo "github.com/x-xyz/goapi/stores/apecoinstaking/repository"
	apecoinstakingUseCase "github.com/x-xyz/goapi/stores/apecoinstaking/usecase"
	chainlink_usecase "github.com/x-xyz/goapi/stores/chainlink/usecase"
	collection_reposiroty "github.com/x-xyz/goapi/stores/collection/repository"
	collection_usecase "github.com/x-xyz/goapi/stores/collection/usecase"
	erc1155_repository "github.com/x-xyz/goapi/stores/erc1155/repository"
	openseadata_repository "github.com/x-xyz/goapi/stores/openseadata/repository"
	openseadata_usecase "github.com/x-xyz/goapi/stores/openseadata/usecase"
	order_repository "github.com/x-xyz/goapi/stores/order/repository"
	paytoken_repository "github.com/x-xyz/goapi/stores/paytoken/repository"
	token_repository "github.com/x-xyz/goapi/stores/token/repository"
	token_usecase "github.com/x-xyz/goapi/stores/token/usecase"
	webresource_repository "github.com/x-xyz/goapi/stores/web_resource/repository"
	webresource_usecase "github.com/x-xyz/goapi/stores/web_resource/usecase"
)

func init() {
	viper.SetConfigType("yaml")
	viper.SetConfigFile(`infra/configs/nft-indexer/config.yaml`)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	if viper.GetBool(`debug`) {
		log.Log().Info("Service RUN on DEBUG mode")
	}
}

func main() {
	// start server to pass cloud run health check
	startEchoServer()

	ctx, cancel := bCtx.WithCancel(bCtx.Background())

	ipfsApiUrl := viper.GetString("ipfs.api")
	ipfsTimeout := viper.GetDuration("ipfs.timeout")
	cloudStorageBucket := viper.GetString("cloud-storage.bucket")
	cloudStorageUrl := viper.GetString("cloud-storage.url")
	cloudStorageTimeout := viper.GetDuration("cloud-storage.timeout")
	httpTimeout := viper.GetDuration("http.timeout")
	indexerRetryLimit := viper.GetInt("indexer.retryLimit")
	indexerBatch := viper.GetInt("indexer.batch")
	indexerWorkers := viper.GetInt("indexer.workers")
	indexerInterval := viper.GetDuration("indexer.interval")
	indexerStatInterval := viper.GetDuration("indexer.statInterval")
	osIndexerEnable := viper.GetBool("openseaIndexer.enable")
	osIndexerInterval := viper.GetDuration("openseaIndexer.interval")
	osIndexerRetryLimit := viper.GetInt("openseaIndexer.retryLimit")
	osIndexerBackoffStartD := viper.GetDuration("openseaIndexer.backoffStartDuration")
	osIndexerBackoffLimitD := viper.GetDuration("openseaIndexer.backoffLimitDuration")
	osIndexerApikey := viper.GetString("openseaIndexer.apikey")
	osEventIndexerEnable := viper.GetBool("openseaEventIndexer.enable")
	osEventIndexerInterval := viper.GetDuration("openseaEventIndexer.interval")
	osEventIndexerRetryLimit := viper.GetInt("openseaEventIndexer.retryLimit")
	osEventIndexerBackoffStartD := viper.GetDuration("openseaEventIndexer.backoffStartDuration")
	osEventIndexerBackoffLimitD := viper.GetDuration("openseaEventIndexer.backoffLimitDuration")
	osEventIndexerApikey := viper.GetString("openseaEventIndexer.apikey")
	thumborUrl := viper.GetString("thumbor.url")

	ctx.WithFields(log.Fields{
		"ipfs.api":              ipfsApiUrl,
		"ipfs.timeout":          ipfsTimeout,
		"cloud-storage.bucket":  cloudStorageBucket,
		"cloud-storage.url":     cloudStorageUrl,
		"cloud-storage.timeout": cloudStorageTimeout,
		"http.timeout":          httpTimeout,
		"indexer.retryLimit":    indexerRetryLimit,
		"indexer.batch":         indexerBatch,
		"indexer.workers":       indexerWorkers,
		"indexer.interval":      indexerInterval,
		"indexer.statInterval":  indexerStatInterval,
	}).Info("config")

	ctx.Info("init mongo")
	q := initMongo()
	ipfsShell := ipfsapi.NewShell(ipfsApiUrl)
	networks := viper.Sub("networks")
	keys := networks.AllSettings()
	rpcs := make(map[int32]string)
	archiveRpcs := make(map[int32]string)
	for k := range keys {
		chainId := networks.GetInt32(fmt.Sprintf("%s.chainId", k))
		rpcUrl := networks.GetString(fmt.Sprintf("%s.rpcUrl", k))
		rpcs[chainId] = rpcUrl
		archiveRpcUrl := networks.GetString(fmt.Sprintf("%s.archiveRpcUrl", k))
		archiveRpcs[chainId] = archiveRpcUrl
	}
	chainService, err := chain.NewClient(ctx, &chain.ClientCfg{
		RpcUrls:        rpcs,
		ArchiveRpcUrls: archiveRpcs,
	})
	if err != nil {
		ctx.WithField("err", err).Warn("chainService started with error")
	}
	httpClient := http.Client{}
	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		ctx.WithField("err", err).Panic("storage.NewClient failed")
	}
	openseaClient := opensea.NewClient(&opensea.ClientCfg{
		HttpClient: http.Client{},
		Timeout:    httpTimeout,
		Apikey:     osIndexerApikey,
	})
	openseaEventClient := opensea.NewClient(&opensea.ClientCfg{
		HttpClient: http.Client{},
		Timeout:    httpTimeout,
		Apikey:     osEventIndexerApikey,
	})
	chainlinkService := chainlink_service.New(chainService)
	errCh := make(chan error, 10)
	var indexers []*nft_indexer.NftIndexer

	// repos
	httpRepo := webresource_repository.NewHttpReaderRepo(httpClient, httpTimeout, nil)
	datauriRepo := webresource_repository.NewDataUriReaderRepo()
	aruriRepo := webresource_repository.NewArReaderRepo(httpClient, httpTimeout, nil)
	ipfsNodeRepo := webresource_repository.NewIpfsNodeApiReaderRepo(ipfsShell, ipfsTimeout)
	cloudStorageRepo, err := webresource_repository.NewCloudStorageWriterRepo(&webresource_repository.CloudStorageWriterRepoCfg{
		Timeout:    cloudStorageTimeout,
		Client:     storageClient,
		BucketName: cloudStorageBucket,
		Url:        cloudStorageUrl,
	})
	if err != nil {
		ctx.WithField("err", err).Panic("NewCloudStorageWriterRepo failed")
	}
	nftitemRepo := token_repository.NewNftItem(q, nil)
	collectionRepo := collection_reposiroty.NewCollection(q)
	erc1155HoldingRepo := erc1155_repository.NewHoldingRepo(q)
	openseaDataRepo := openseadata_repository.NewOpenseaDataRepo(q)
	floorPriceHistoryRepo := collection_reposiroty.NewFloorPriceHistoryRepo(q)
	paytokenRepo := paytoken_repository.NewPayTokenRepo(q)
	orderItemRepo := order_repository.NewOrderItemRepo(q)
	apecoinStakingRepo := apecoinstakingRepo.New(q)

	// usecases
	webResourceUseCase := webresource_usecase.NewWebResourceUseCase(&webresource_usecase.WebResourceUseCaseCfg{
		HttpReader:         httpRepo,
		IpfsReader:         ipfsNodeRepo,
		DataUriReader:      datauriRepo,
		ArUriReader:        aruriRepo,
		CloudStorageWriter: cloudStorageRepo,
	})
	chainlink := chainlink_usecase.New(chainlinkService, paytokenRepo)
	tokenUseCase := token_usecase.New(&token_usecase.TokenUseCaseCfg{
		NftitemRepo: nftitemRepo,
	})
	collectionUseCase := collection_usecase.NewCollection(&collection_usecase.CollectionUseCaseCfg{
		CollectionRepo:        collectionRepo,
		Erc1155holdingRepo:    erc1155HoldingRepo,
		NftitemRepo:           nftitemRepo,
		ChainlinkUC:           chainlink,
		FloorPriceHistoryRepo: floorPriceHistoryRepo,
		OrderItemRepo:         orderItemRepo,
	})
	openseaDataUseCase := openseadata_usecase.NewOpenseaUseCase(openseaDataRepo)
	activityHistoryRepo := account_repository.NewActivityHistoryRepo(q)
	activityHistoryUseCase := account_usecase.NewActivityHistoryUsecase(activityHistoryRepo)
	apecoinStakingUseCase := apecoinstakingUseCase.New(apecoinStakingRepo)

	indexerStates := []nftitem.IndexerState{
		nftitem.IndexerStateHasTokenURI,
		nftitem.IndexerStateHasTokenURIRefreshing,
		nftitem.IndexerStateHasImageURL,
		nftitem.IndexerStateHasHostedImage,
		nftitem.IndexerStateParsingAttributes,
		nftitem.IndexerStateFetchingAnimation,
		nftitem.IndexerStateBeforeMigrateMimeType,
	}
	parserSelector := metadata_parser.NewSelector(metadata_parser.NewDefaultParser())
	metadata_parser.InitializeSelector(parserSelector, webResourceUseCase, apecoinStakingUseCase)
	animationParserSelector := animation_url_parser.NewSelector(animation_url_parser.NewDefaultParser())
	for _, s := range indexerStates {
		indexer := nft_indexer.NewNftIndexer(&nft_indexer.NftIndexerCfg{
			WebResourceUC:              webResourceUseCase,
			TokenUC:                    tokenUseCase,
			TargetState:                s,
			RetryLimit:                 indexerRetryLimit,
			Batch:                      indexerBatch,
			Workers:                    indexerWorkers,
			Interval:                   indexerInterval,
			ErrorCh:                    errCh,
			ThumborUrl:                 thumborUrl,
			ParserSelector:             parserSelector,
			AnimationUrlParserSelector: animationParserSelector,
		})
		indexers = append(indexers, indexer)
	}

	for _, i := range indexers {
		i.Start(ctx)
	}

	statUpdater := nft_indexer.
		NewStatUpdater(collectionUseCase, errCh).
		SetInterval(indexerStatInterval)
	statUpdater.Start(ctx)

	osIndexer := nft_indexer.NewOpenseaDataIndexer(&nft_indexer.OpenseaDataIndexerCfg{
		Collection:    collectionUseCase,
		OpenseaData:   openseaDataUseCase,
		OpenseaClient: openseaClient,
		RetryLimit:    osIndexerRetryLimit,
		Backoff:       backoff.NewExponential(osIndexerBackoffStartD, osIndexerBackoffLimitD),
		Interval:      osIndexerInterval,
		ErrorCh:       errCh,
	})
	if osIndexerEnable {
		osIndexer.Start(ctx)
	}

	osEventIndexer := nft_indexer.NewOpenseaEventIndexer(&nft_indexer.OpenseaEventIndexerCfg{
		Collection:             collectionUseCase,
		ActivityHistoryUsecase: activityHistoryUseCase,
		OpenseaClient:          openseaEventClient,
		RetryLimit:             osEventIndexerRetryLimit,
		Backoff:                backoff.NewExponential(osEventIndexerBackoffStartD, osEventIndexerBackoffLimitD),
		Interval:               osEventIndexerInterval,
		ErrorCh:                errCh,
	})
	if osEventIndexerEnable {
		osEventIndexer.Start(ctx)
	}

	// wait for first error
	err = <-errCh
	ctx.WithField("err", err).Error("indexer error")
	go func() {
		for range errCh {
		}
	}()
	cancel()
	for _, i := range indexers {
		i.Wait()
	}

	statUpdater.Wait()
	if osIndexerEnable {
		osIndexer.Wait()
	}
	if osEventIndexerEnable {
		osEventIndexer.Wait()
	}
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
