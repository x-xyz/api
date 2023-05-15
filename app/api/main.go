package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"

	"github.com/go-playground/validator/v10"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/base/database/redisclient"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/base/metrics"
	pricefomatter "github.com/x-xyz/goapi/base/price_fomatter"
	bValidator "github.com/x-xyz/goapi/base/validator"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/order"
	mmiddleware "github.com/x-xyz/goapi/middleware"
	"github.com/x-xyz/goapi/service/chain"
	"github.com/x-xyz/goapi/service/chain/contract"
	chainlink_service "github.com/x-xyz/goapi/service/chainlink"
	"github.com/x-xyz/goapi/service/coingecko"
	"github.com/x-xyz/goapi/service/ens"
	"github.com/x-xyz/goapi/service/hyype"
	"github.com/x-xyz/goapi/service/opensea"
	"github.com/x-xyz/goapi/service/pinata"
	"github.com/x-xyz/goapi/service/query"
	"github.com/x-xyz/goapi/service/redis"
	account_delivery "github.com/x-xyz/goapi/stores/account/delivery/http"
	account_repository "github.com/x-xyz/goapi/stores/account/repository"
	account_usecase "github.com/x-xyz/goapi/stores/account/usecase"
	airdrop_delivery "github.com/x-xyz/goapi/stores/airdrop/delivery/http"
	airdrop_repository "github.com/x-xyz/goapi/stores/airdrop/repository"
	airdrop_usecase "github.com/x-xyz/goapi/stores/airdrop/usecase"
	auth_delivery "github.com/x-xyz/goapi/stores/auth/delivery/http"
	auth_middleware "github.com/x-xyz/goapi/stores/auth/delivery/http/middleware"
	auth_usecase "github.com/x-xyz/goapi/stores/auth/usecase"
	chainlink_usecase "github.com/x-xyz/goapi/stores/chainlink/usecase"
	coin_delivery "github.com/x-xyz/goapi/stores/coin/delivery/http"
	collection_delivery "github.com/x-xyz/goapi/stores/collection/delivery/http"
	collection_repository "github.com/x-xyz/goapi/stores/collection/repository"
	collection_usecase "github.com/x-xyz/goapi/stores/collection/usecase"
	coll_promotion_delivery "github.com/x-xyz/goapi/stores/collection_promotion/delivery/http"
	coll_promotion_repository "github.com/x-xyz/goapi/stores/collection_promotion/repository"
	coll_promotion_usecase "github.com/x-xyz/goapi/stores/collection_promotion/usecase"
	ens_delivery "github.com/x-xyz/goapi/stores/ens/delivery/http"
	erc1155Repository "github.com/x-xyz/goapi/stores/erc1155/repository"
	external_listing_delivery "github.com/x-xyz/goapi/stores/external_listing/delivery/http"
	external_listing_repository "github.com/x-xyz/goapi/stores/external_listing/repository"
	external_listing_usecase "github.com/x-xyz/goapi/stores/external_listing/usecase"
	file_usecase "github.com/x-xyz/goapi/stores/file/usecase"
	hc_delivery "github.com/x-xyz/goapi/stores/healthcheck/delivery/http"
	hc_repo "github.com/x-xyz/goapi/stores/healthcheck/repository"
	hc_usecase "github.com/x-xyz/goapi/stores/healthcheck/usecase"
	ip_delivery "github.com/x-xyz/goapi/stores/ip/delivery/http"
	ip_repository "github.com/x-xyz/goapi/stores/ip/repository"
	ip_usecase "github.com/x-xyz/goapi/stores/ip/usecase"
	moderator_delivery "github.com/x-xyz/goapi/stores/moderator/delivery/http"
	moderator_repository "github.com/x-xyz/goapi/stores/moderator/repository"
	moderator_usecase "github.com/x-xyz/goapi/stores/moderator/usecase"
	openseadata_repository "github.com/x-xyz/goapi/stores/openseadata/repository"
	order_repository "github.com/x-xyz/goapi/stores/order/repository"
	order_usecase "github.com/x-xyz/goapi/stores/order/usecase"
	paytoken_repository "github.com/x-xyz/goapi/stores/paytoken/repository"
	promotion_delivery "github.com/x-xyz/goapi/stores/promotion/delivery/http"
	promotion_repository "github.com/x-xyz/goapi/stores/promotion/repository"
	promotion_usecase "github.com/x-xyz/goapi/stores/promotion/usecase"
	relationship_repository "github.com/x-xyz/goapi/stores/relationship/repository"
	relationship_usecase "github.com/x-xyz/goapi/stores/relationship/usecase"
	search_delivery "github.com/x-xyz/goapi/stores/search/delivery/http"
	search_usecase "github.com/x-xyz/goapi/stores/search/usecase"
	statistic_delivery "github.com/x-xyz/goapi/stores/statistic/delivery/http"
	statistics_repository "github.com/x-xyz/goapi/stores/statistic/repository"
	statistics_usecase "github.com/x-xyz/goapi/stores/statistic/usecase"
	token_delivery "github.com/x-xyz/goapi/stores/token/delivery/http"
	token_repository "github.com/x-xyz/goapi/stores/token/repository"
	token_usecase "github.com/x-xyz/goapi/stores/token/usecase"
	twelvefold_delivery "github.com/x-xyz/goapi/stores/twelvefold/delivery/http"
	twelvefold_repository "github.com/x-xyz/goapi/stores/twelvefold/repository"
	twelvefold_usecase "github.com/x-xyz/goapi/stores/twelvefold/usecase"
	vex_delivery "github.com/x-xyz/goapi/stores/vex/delivery/http"
	vex_repository "github.com/x-xyz/goapi/stores/vex/repository"
	vex_usecase "github.com/x-xyz/goapi/stores/vex/usecase"

	echoSwagger "github.com/swaggo/echo-swagger"

	_ "github.com/x-xyz/goapi/app/api/docs"
)

func init() {
	viper.SetConfigType("yaml")
	viper.SetConfigFile(`infra/configs/config.yaml`)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	if viper.GetBool(`debug`) {
		log.Log().Info("Service RUN on DEBUG mode")
	}
}

//	@title			X Marketplace API
//	@version		1.0
//	@description	API Document for X Marketplace.

// main
//
//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						Authorization
//	@description				retrive token from #/auth/post_auth_sign and apply with `bearer {token}`
func main() {
	// init echo
	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{}))
	e.Use(middleware.RequestID())
	middL := mmiddleware.InitMiddleware()
	e.Use(middL.ResponseLogger())
	e.Use(middL.AddContext())
	e.Use(middleware.CORS())
	e.Validator = bValidator.NewCustomValidator(validator.New())

	context := ctx.Background()

	// init mongo client
	context.Info("init mongo")
	uri := viper.GetString("mongo.uri")
	authDBName := viper.GetString("mongo.authDBName")
	dbName := viper.GetString("mongo.dbName")
	enableSSL := viper.GetBool("mongo.enableSSL")
	checkIndex := viper.GetBool("mongo.checkIndex")
	mongoClient := mongoclient.MustConnectMongoClient(uri, authDBName, dbName, enableSSL, true, 2)
	q := query.New(mongoClient, checkIndex)

	// init Redis service
	context.Info("init redis cache")
	redisCacheName := viper.GetString("redis_cache.name")
	redisCacheURI := viper.GetString("redis_cache.uri")
	redisCachePwd := viper.GetString("redis_cache.password")
	redisCachePoolMultiplier := viper.GetFloat64("redis_cache.poolMultiplier")
	redisCachePool := redisclient.MustConnectRedis(redisCacheURI, redisCachePwd, redisclient.RedisParam{
		PoolMultiplier: redisCachePoolMultiplier,
		Retry:          true,
	})
	redisCache := redis.New(redisCacheName, metrics.New(redisCacheName), &redis.Pools{
		Src: redisCachePool,
	})

	mmiddleware.SetupCache(redisCache)

	pinata := pinata.New(viper.GetString("pinata.apiKey"), viper.GetString("pinata.apiSecret"))

	hyypeApiKey := viper.GetString("hyype.apikey")
	hyypeTimeout := viper.GetDuration("hyype.timeout")
	hyypeClient := hyype.NewClient(&hyype.ClientCfg{
		HttpClient: http.Client{},
		Timeout:    hyypeTimeout,
		Apikey:     hyypeApiKey,
	})

	exchanges := viper.Sub("exchanges")
	keys := exchanges.AllSettings()
	exchangeCfgs := make(map[domain.ChainId]order.ExchangeCfg)
	for k := range keys {
		chainId := domain.ChainId(exchanges.GetInt(fmt.Sprintf("%s.chainId", k)))
		exchangeAddr := exchanges.GetString(fmt.Sprintf("%s.exchange", k))
		exchangeCfgs[chainId] = order.ExchangeCfg{
			Address:    domain.Address(exchangeAddr).ToLower(),
			Strategies: make(map[domain.Address]order.Strategy),
		}
		strategies := exchanges.GetStringMapString(fmt.Sprintf("%s.strategies", k))
		for addr, name := range strategies {
			exchangeCfgs[chainId].Strategies[domain.Address(addr).ToLower()] = order.ToStrategy(name)
		}
	}

	cacheDuration := viper.GetFloat64("externalListing.cacheDuration")

	// init chain service
	networks := viper.Sub("networks")
	keys = networks.AllSettings()
	rpcs := make(map[int32]string)
	archiveRpcs := make(map[int32]string)
	for k := range keys {
		chainId := networks.GetInt32(fmt.Sprintf("%s.chainId", k))
		rpcUrl := networks.GetString(fmt.Sprintf("%s.rpcUrl", k))
		rpcs[chainId] = rpcUrl
		archiveRpcUrl := networks.GetString(fmt.Sprintf("%s.archiveRpcUrl", k))
		archiveRpcs[chainId] = archiveRpcUrl
	}
	chainService, err := chain.NewClient(context, &chain.ClientCfg{
		RpcUrls:        rpcs,
		ArchiveRpcUrls: archiveRpcs,
	})
	if err != nil {
		context.WithField("err", err).Warn("chainService started with error")
	}
	erc721Service := contract.NewErc721(chainService)
	erc1155Service := contract.NewErc1155(chainService)
	erc1271Service := contract.NewErc1271(chainService)
	chainlinkService := chainlink_service.New(chainService)
	coinGecko := coingecko.NewClient(&coingecko.ClientCfg{
		HttpClient: http.Client{},
		Timeout:    10 * time.Second,
	})
	httpTimeout := viper.GetDuration("http.timeout")
	openseaApiKey := viper.GetString("opensea.apikey")
	openseaClient := opensea.NewClient(&opensea.ClientCfg{
		HttpClient: http.Client{},
		Timeout:    httpTimeout,
		Apikey:     openseaApiKey,
	})
	// ens on ethereum
	ensService := ens.New(rpcs[1], redisCache)

	// construct repository, usecase and delivery
	hcRepo := hc_repo.New(mongoClient, redisCache)
	nftitemRepo := token_repository.NewNftItem(q, redisCache)
	unlockableRepo := token_repository.NewUnlockable(q)
	collectionRepo := collection_repository.NewCollection(q)
	registrationRepo := collection_repository.NewRegistration(q)
	accountRepo := account_repository.New(q, redisCache)
	nsRepo := account_repository.NewNotificationSettingsRepo(q)
	moderatorRepo := moderator_repository.New(q)
	followRepo := relationship_repository.NewFollow(q)
	likeRepo := relationship_repository.NewLike(q)
	collectionLikeRepo := relationship_repository.NewCollectionLike(q)
	erc721Repo := collection_repository.NewErc721Contract(q)
	erc1155Repo := erc1155Repository.NewContractRepo(q)
	erc1155HoldingRepo := erc1155Repository.NewHoldingRepo(q)
	airdropRepo := airdrop_repository.NewAirdropRepo(q)
	proofRepo := airdrop_repository.NewProofRepo(q)
	openseaDataRepo := openseadata_repository.NewOpenseaDataRepo(q)
	paytokenRepo := paytoken_repository.NewPayTokenRepo(q)
	tradingVolumeRepo := collection_repository.NewTradingVolumeRepo(q)
	activityRepo := account_repository.NewActivityHistoryRepo(q)
	vexRepo := vex_repository.NewVexFeeDistributionHistoryRepo(q)
	folderRepo := account_repository.NewFolderRepo(q)
	folderRelationRepo := account_repository.NewFolderNftRelationshipRepo(q)
	floorPriceHistoryRepo := collection_repository.NewFloorPriceHistoryRepo(q)
	promotionRepo := promotion_repository.NewPromotion(q)
	collPromotionRepo := coll_promotion_repository.NewCollPromotion(q)
	listingRecordRepo := airdrop_repository.NewListingRecordRepo(q)
	orderItemRepo := order_repository.NewOrderItemRepo(q)
	orderRepo := order_repository.NewOrderRepo(q)
	orderNonceRepo := account_repository.NewOrderNonceRepo(q)
	externalListingRepo := external_listing_repository.NewExternalListingRepo(q)
	statisticRepo := statistics_repository.New(q)
	ipRepo := ip_repository.New(q)
	twelvefoldRepo := twelvefold_repository.NewTwelvefoldRepo(q)

	chainlink := chainlink_usecase.New(chainlinkService, paytokenRepo)
	priceFormatter := pricefomatter.NewPriceFormatter(&pricefomatter.PriceFormatterCfg{
		Paytoken:  paytokenRepo,
		Chainlink: chainlink,
		CoinGecko: coinGecko,
	})
	file := file_usecase.New(pinata)
	hc := hc_usecase.New(hcRepo)
	promotionUsecase := promotion_usecase.NewPromotion(promotionRepo)
	collPromotionUsecase := coll_promotion_usecase.NewCollPromotion(&coll_promotion_usecase.CollPromotionCfg{
		CollPromotionRepo: collPromotionRepo,
		PromotionUsecase:  promotionUsecase,
		ListingRecordRepo: listingRecordRepo,
		OpenseaDataRepo:   openseaDataRepo,
	})
	token := token_usecase.New(&token_usecase.TokenUseCaseCfg{
		LikeRepo:           likeRepo,
		NftitemRepo:        nftitemRepo,
		CollectionRepo:     collectionRepo,
		UnlockableRepo:     unlockableRepo,
		FileUC:             file,
		IpfsUri:            viper.GetString("ipfsUri"),
		ActivityRepo:       activityRepo,
		FolderRelationRepo: folderRelationRepo,
		Erc1155HoldingRepo: erc1155HoldingRepo,
		OrderItemRepo:      orderItemRepo,
		Redis:              redisCache,
	})
	collection := collection_usecase.NewCollection(&collection_usecase.CollectionUseCaseCfg{
		CollectionRepo:        collectionRepo,
		RegistrationRepo:      registrationRepo,
		Erc721contractRepo:    erc721Repo,
		Erc1155contractRepo:   erc1155Repo,
		Erc1155holdingRepo:    erc1155HoldingRepo,
		FileUC:                file,
		Erc721ChainService:    erc721Service,
		Erc1155ChainService:   erc1155Service,
		NftitemRepo:           nftitemRepo,
		OpenseaDataRepo:       openseaDataRepo,
		ChainlinkUC:           chainlink,
		FloorPriceHistoryRepo: floorPriceHistoryRepo,
		OrderItemRepo:         orderItemRepo,
		ActivityHistoryRepo:   activityRepo,
		LikeRepo:              likeRepo,
		PromotedCollectionsUC: collPromotionUsecase,
		TokenUC:               token,
	})
	follow := relationship_usecase.NewFollow(followRepo)
	like := relationship_usecase.NewLike(likeRepo, nftitemRepo)
	collectionLike := relationship_usecase.NewCollectionLike(collectionLikeRepo, collectionRepo)
	moderator := moderator_usecase.New(moderatorRepo)
	folderUsecase := account_usecase.NewFolderUsecase(folderRepo, folderRelationRepo, nftitemRepo, collectionRepo, floorPriceHistoryRepo, coinGecko, erc1155HoldingRepo)
	account := account_usecase.New(&account_usecase.AccountUseCaseCfg{
		Repo:                    accountRepo,
		NotificationSettingRepo: nsRepo,
		FollowUC:                follow,
		ModeratorUC:             moderator,
		FileUC:                  file,
		NftitemRepo:             nftitemRepo,
		TokenUC:                 token,
		LikeUC:                  like,
		SignatureMsg:            viper.GetString("auth.signatureMsg"),
		CollectionUC:            collection,
		ActivityRepo:            activityRepo,
		FolderUC:                folderUsecase,
	})
	auth := auth_usecase.New(viper.GetString("auth.jwtSecret"), account)
	search := search_usecase.New(q)
	airdrop := airdrop_usecase.NewAirdropUseCase(airdropRepo)
	proof := airdrop_usecase.NewProofUseCase(proofRepo)
	tradingVolume := collection_usecase.NewTradingVolumeUseCase(tradingVolumeRepo, chainlink)
	vex := vex_usecase.NewVexFeeDistrubutionHistoryUseCase(vexRepo)
	orderNonce := account_usecase.NewOrderNonceUseCase(orderNonceRepo)
	order := order_usecase.New(&order_usecase.OrderUseCaseCfg{
		ExchangeCfgs:        exchangeCfgs,
		OrderRepo:           orderRepo,
		OrderItemRepo:       orderItemRepo,
		NftitemRepo:         nftitemRepo,
		Erc1155HoldingRepo:  erc1155HoldingRepo,
		AccountRepo:         accountRepo,
		PaytokenRepo:        paytokenRepo,
		PriceFormatter:      priceFormatter,
		OrderNonceUC:        orderNonce,
		TokenUC:             token,
		Erc1271:             erc1271Service,
		ActivityHistoryRepo: activityRepo,
	})
	externalListingUsecase := external_listing_usecase.New(openseaClient, externalListingRepo, priceFormatter)
	statisticUsecase := statistics_usecase.New(statisticRepo)
	ipUseCase := ip_usecase.New(ipRepo, nftitemRepo)
	twelvefoldUseCase := twelvefold_usecase.NewTwelvefoldUseCase(twelvefoldRepo)

	adminAddresses := viper.GetStringSlice("admin.addresses")
	auth_middleware := auth_middleware.New(auth, moderator, adminAddresses)

	hc_delivery.New(e, hc)
	auth_delivery.New(e, auth, viper.GetString("auth.signatureMsg"))
	account_delivery.New(e, account, like, folderUsecase, collection, auth_middleware, orderNonce)
	token_delivery.New(e, token, like, account, folderUsecase, order, auth_middleware, hyypeClient)
	collection_delivery.New(e, account, collection, auth_middleware, collectionLike, tradingVolume)
	moderator_delivery.New(e, moderator, account, auth_middleware)
	search_delivery.New(e, search)
	airdrop_delivery.New(e, airdrop, proof)
	vex_delivery.New(e, vex)
	promotion_delivery.New(e, promotionUsecase)
	coll_promotion_delivery.New(e, collPromotionUsecase, collection)
	coin_delivery.New(e, coinGecko)
	external_listing_delivery.New(e, externalListingUsecase, collection, cacheDuration)
	statistic_delivery.New(e, statisticUsecase)
	ip_delivery.New(e, ipUseCase, account, auth_middleware)
	ens_delivery.New(e, ensService)
	twelvefold_delivery.New(e, twelvefoldUseCase)

	e.GET("/check", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"address": c.Get("address").(string),
		})
	}, auth_middleware.Auth())

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	go func() {
		if err := e.Start(viper.GetString("server.address")); err != nil && err != http.ErrServerClosed {
			log.Log().WithField("err", err).Error("shutting down the server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	sig := <-quit
	log.Log().WithField("signal", sig).Info("received signal")
	ctx, cancel := ctx.WithTimeout(context, 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		log.Log().WithField("err", err).Error("shutting down the server")
	} else {
		log.Log().Info("shutdown server successfully")
	}
}
