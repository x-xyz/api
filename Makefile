IMAGE_REPO = asia-east1-docker.pkg.dev/gcp_project/x
GOAPI_IMAGE = $(IMAGE_REPO)/goapi
TRACKER_IMAGE = $(IMAGE_REPO)/tracker
GO_IPFS_IMAGE = $(IMAGE_REPO)/go-ipfs
NFT_INDEXER_IMAGE = $(IMAGE_REPO)/nft-indexer
BUILD_ID = $(shell git rev-parse HEAD)
DOCKERFILE = infra/docker

test:
	go test -v -cover -covermode=atomic ./...

unittest:
	go test -short  ./...

prepare-go-ipfs: check-tag context
	docker pull ipfs/go-ipfs:$(TAG)
	docker tag ipfs/go-ipfs:$(TAG) $(GO_IPFS_IMAGE):$(TAG)
	docker push $(GO_IPFS_IMAGE):$(TAG)

build: build-doc
	docker build --platform linux/amd64 -f $(DOCKERFILE)/Dockerfile.goapi -t $(GOAPI_IMAGE):$(BUILD_ID) .
	docker push $(GOAPI_IMAGE):$(BUILD_ID)

	docker build --platform linux/amd64 -f $(DOCKERFILE)/Dockerfile.tracker -t $(TRACKER_IMAGE):$(BUILD_ID) .
	docker push $(TRACKER_IMAGE):$(BUILD_ID)

build-doc:
	swag fmt
	swag init -g ./app/api/main.go -o ./app/api/docs

build-nft-indexer:
	docker build --platform linux/amd64 -f $(DOCKERFILE)/Dockerfile.nft-indexer -t $(NFT_INDEXER_IMAGE):$(BUILD_ID) .
	docker push $(NFT_INDEXER_IMAGE):$(BUILD_ID)

run:
	docker-compose up --build -d

stop:
	docker-compose down

lint-prepare:
	@echo "Installing golangci-lint" 
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s latest

lint:
	./bin/golangci-lint run ./...

context:
	@echo ">> Switching gcloud project to 'gcp_project'"
	gcloud config set project gcp_project

deploy-dev: check-tag context
	# update secret
	gcloud secrets versions add dev-config --data-file="./infra/configs/dev-config.yaml"
	gcloud secrets versions add dev-tracker-config --data-file="./infra/configs/tracker/dev-config.yaml"

	# update image tag
	cd infra/cloudrun/overlays/dev && \
	kustomize edit set image $(GOAPI_IMAGE)=$(GOAPI_IMAGE):$(TAG) && \
	kustomize edit set image $(TRACKER_IMAGE)=$(TRACKER_IMAGE):$(TAG) && cd -

	# generate deployment yamls
	kustomize build infra/cloudrun/overlays/dev | kubesplit -q -i - -o infra/cloudrun/deploy/tmp
	gcloud beta run services replace infra/cloudrun/deploy/tmp/30--service--xxyz-goapi-dev.yml --region asia-east1 --platform managed
	gcloud beta run services replace infra/cloudrun/deploy/tmp/30--service--xxyz-tracker-eth-dev.yml --region asia-east1 --platform managed
	gcloud beta run services replace infra/cloudrun/deploy/tmp/30--service--xxyz-tracker-eth-mainnet-dev.yml --region asia-east1 --platform managed

	make clean

deploy-nft-indexer-dev: check-tag context
	gcloud secrets versions add dev-nft-indexer-config --data-file="./infra/configs/nft-indexer/dev-config.yaml"

	cd infra/cloudrun/overlays/dev && \
	kustomize edit set image $(NFT_INDEXER_IMAGE)=$(NFT_INDEXER_IMAGE):$(TAG) && cd -

	kustomize build infra/cloudrun/overlays/dev | kubesplit -q -i - -o infra/cloudrun/deploy/tmp
	gcloud beta run services replace infra/cloudrun/deploy/tmp/30--service--xxyz-nft-indexer-dev.yml --region asia-east1 --platform managed

	make clean

deploy-prod: check-tag context
	# update secret
	gcloud secrets versions add prod-config --data-file="./infra/configs/prod-config.yaml"
	gcloud secrets versions add prod-tracker-config --data-file="./infra/configs/tracker/prod-config.yaml"

	# update image tag
	cd infra/cloudrun/overlays/prod && \
	kustomize edit set image $(GOAPI_IMAGE)=$(GOAPI_IMAGE):$(TAG) && \
	kustomize edit set image $(TRACKER_IMAGE)=$(TRACKER_IMAGE):$(TAG) && cd -

	# generate deployment yamls
	kustomize build infra/cloudrun/overlays/prod | kubesplit -q -i - -o infra/cloudrun/deploy/tmp
	gcloud beta run services replace infra/cloudrun/deploy/tmp/30--service--xxyz-goapi-prod.yml --region asia-east1 --platform managed
	gcloud beta run services replace infra/cloudrun/deploy/tmp/30--service--xxyz-tracker-eth-prod.yml --region asia-east1 --platform managed

	make clean

deploy-nft-indexer-prod: check-tag context
	gcloud secrets versions add prod-nft-indexer-config --data-file="./infra/configs/nft-indexer/prod-config.yaml"

	cd infra/cloudrun/overlays/prod && \
	kustomize edit set image $(NFT_INDEXER_IMAGE)=$(NFT_INDEXER_IMAGE):$(TAG) && cd -

	kustomize build infra/cloudrun/overlays/prod | kubesplit -q -i - -o infra/cloudrun/deploy/tmp
	gcloud beta run services replace infra/cloudrun/deploy/tmp/30--service--xxyz-nft-indexer-prod.yml --region asia-east1 --platform managed

	make clean

deploy-thumbor: context
	gcloud beta run services replace infra/cloudrun/base/thumbor.yaml --region asia-east1 --platform managed

	make clean

create-cluster-dev: context
	gcloud container clusters create xxyz-dev --num-nodes=1 --zone asia-east1-a --network=vpc-dev --machine-type=e2-small

create-cluster-prod: context
	gcloud container clusters create xxyz-prod --num-nodes=1 --zone asia-east1-a --network=vpc-prod --machine-type=e2-medium

deploy-go-ipfs-dev: check-tag context get-cluster-cred-dev
	cd infra/gke/overlays/dev && \
	kustomize edit set image $(GO_IPFS_IMAGE)=$(GO_IPFS_IMAGE):$(TAG) && cd -

	kustomize build infra/gke/overlays/dev | kubectl apply -f -

	make clean

deploy-go-ipfs-prod: check-tag context get-cluster-cred-prod
	cd infra/gke/overlays/prod && \
	kustomize edit set image $(GO_IPFS_IMAGE)=$(GO_IPFS_IMAGE):$(TAG) && cd -

	kustomize build infra/gke/overlays/prod | kubectl apply -f -

	make clean

get-cluster-cred-dev: context
	gcloud container clusters get-credentials xxyz-dev --zone asia-east1-a

get-cluster-cred-prod: context
	gcloud container clusters get-credentials xxyz-prod --zone asia-east1-a

get-config-dev:
	@echo ">>> get dev config" 
	@echo ">>>" 
	gcloud secrets versions access latest --secret=dev-config
	@echo ">>> get dev tracker config" 
	@echo ">>>" 
	gcloud secrets versions access latest --secret=dev-tracker-config
	@echo ">>> get dev nft-indexer config" 
	@echo ">>>" 
	gcloud secrets versions access latest --secret=dev-nft-indexer-config

get-config-prod:
	@echo ">>> get prod config" 
	@echo ">>>" 
	gcloud secrets versions access latest --secret=prod-config
	@echo ">>> get prod tracker config" 
	@echo ">>>" 
	gcloud secrets versions access latest --secret=prod-tracker-config
	@echo ">>> get prod nft-indexer config" 
	@echo ">>>" 
	gcloud secrets versions access latest --secret=prod-nft-indexer-config

deploy-function-weekly-promotion-prod: context
	make -C functions/weekly-promotion deploy-prod

debug-dev:
	@echo ">> Generating 'dev' deployment yamls to 'infra/cloudrun/deploy/debug/xxyz-dev'"
	kustomize build infra/cloudrun/overlays/dev | kubesplit -q -i - -o infra/cloudrun/deploy/debug/xxyz-dev

debug-prod:
	@echo ">> Generating 'prod' deployment yamls to 'infra/cloudrun/deploy/debug/xxyz-prod'"
	kustomize build infra/cloudrun/overlays/prod | kubesplit -q -i - -o infra/cloudrun/deploy/debug/xxyz-prod


check-tag:
ifeq ($(strip $(TAG)),)
	echo "Variable 'TAG' was not set"
	exit 1
endif

clean:
	git checkout infra/cloudrun/overlays/dev/kustomization.yaml
	git checkout infra/cloudrun/overlays/prod/kustomization.yaml
	git checkout infra/gke/overlays/dev/kustomization.yaml
	git checkout infra/gke/overlays/prod/kustomization.yaml
	rm -rf infra/cloudrun/deploy/tmp infra/cloudrun/deploy/debug

.PHONY: clean install unittest build docker run stop vendor lint-prepare lint context deploy-dev deploy-prod debug-dev debug-prod check-tag
