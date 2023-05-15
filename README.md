# goapi

### TODO

- Deployment (local/dev/prod)
- Mongo client
- Logger/Custome Context
- Middlewares (Ex: Auth)
- API/Tracker Implemtation

### Prerequisite
- Golang `1.17.13` (Install [goenv](https://github.com/syndbg/goenv/blob/master/INSTALL.md) to manage local go environment)
- docker
- docker-compose

### Run the Applications

Here is the steps to run it with `docker-compose`

```bash
# Run the application
$ make run

# Execute the call
$ curl localhost:9090/health

# Stop
$ make stop
```

### Run the Testing

```bash
$ make test
```

### Directory Arrangement (See details in [go-clena-arch](https://github.com/bxcodec/go-clean-arch))
This project has 4 Domain layers:
 * Models(Domain) Layer => models, data structure
 * Delivery Layer => HTTP or command line
 * Repository Layer => database or remote service
 * Usecase Layer => business logic

```
├── app                        // application main packages
│   ├── api
│   │   └── main.go
│   └── tracker
│       └── main.go
├── infra                      // infrastructure and config related
│   ├── configs                // application configs, for example: listen prot, database config
│   ├── cloudrun               // GCP cloud run deployment using kustomize
│   ├── database               // database index/migration files
│   └── docker                 // docker related files, like dockerfile, build script
├── docker-compose.yaml
├── domain                     // models, data structure
│   ├── errors.go
│   └── healthcheck.go
├── middleware                 // middleware for echo framework
│   ├── middleware.go
│   └── middleware_test.go
├── service                    // lower layer generic packages
│   └── redis
└── stores                     // different business logic packages
    └── healthcheck
        ├── delivery           // API HTTP handler
        ├── repository         // database access implementation
        └── usecase            // main business logic implementation
```

---

## Cloud Run Deployment

## Instal tools

- [kustomize](https://github.com/kubernetes-sigs/kustomize) >= v4.3.0, < v4.4.0
```
  $ gcloud components install kustomize
```
- [kubesplit](https://github.com/looztra/kubesplit) >= v0.3.2
  - `pip3 install -U --user kubesplit`

### Initialization
1. Deploy the application (see below *Deployment*)
2. (*Optional*) Set IAM policy to allowing public traffic
```
  $ gcloud run services set-iam-policy ${APPLICATION_NAME} infra/cloudrun/policy.yaml --region asia-east1
```

### Deployment
#### Dev
```
  $ make build && make deploy-dev TAG=${IMAGE_TAG}
```

#### Prod

```
  $ make build && make deploy-prod TAG=${IMAGE_TAG}
```

### Debug deployment configuration

```
  $ make debug-dev   ## debug dev yamls
  or
  $ make debug-prod  ## debug prod yamls
```

---

### Architecture
![xxyz drawio (3)](https://user-images.githubusercontent.com/6688990/144755209-3f3ab3b1-ed21-4b3d-887a-d5877c1bc1e7.png)



---

### Tools Used:
In this project, I use some tools listed below. But you can use any simmilar library that have the same purposes. But, well, different library will have different implementation type. Just be creative and use anything that you really need. 

- All libraries listed in [`go.mod`](https://github.com/x-xyz/goapi/blob/master/go.mod) 
- ["github.com/vektra/mockery".](https://github.com/vektra/mockery) To Generate Mocks for testing needs.
- [DB Migration](https://github.com/golang-migrate/migrate/blob/master/database/mongodb/mongodb.go) For updating MongoDB schema

### Echo
- [Echo Guide](https://echo.labstack.com/guide/)

### Go Clean Arch Materials
- [go clean arch](https://github.com/bxcodec/go-clean-arch)
- [Trying Clean Architecture On Golang](https://medium.com/hackernoon/golang-clean-archithecture-efd6d7c43047)
- [Common Anti Patterns In Go Web Applications](https://threedots.tech/post/common-anti-patterns-in-go-web-applications)
