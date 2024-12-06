BINPATH ?= build

BUILD_TIME=$(shell date +%s)
GIT_COMMIT=$(shell git rev-parse HEAD)
VERSION ?= $(shell git tag --points-at HEAD | grep ^v | head -n 1)

LDFLAGS = -ldflags "-X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -X main.Version=$(VERSION)"

.PHONY: all
all: audit test build

.PHONY: audit
audit:
	go list -json -m all | nancy sleuth

.PHONY: build
build:
	go build -tags 'production' $(LDFLAGS) -o $(BINPATH)/dp-upload-service

.PHONY: debug
debug:
	go build -tags 'debug' $(LDFLAGS) -o $(BINPATH)/dp-upload-service
	HUMAN_LOG=1 DEBUG=1 $(BINPATH)/dp-upload-service

.PHONY: test
test:
	go test -count=1 -race -cover ./...

.PHONY: docker-test-component
docker-test-component:
	docker-compose  -f docker-compose-services.yml -f docker-compose.yml down
	docker build -f Dockerfile . -t template_test --target=test
	docker-compose  -f docker-compose-services.yml -f docker-compose.yml up -d
	docker-compose  -f docker-compose-services.yml -f docker-compose.yml exec -T http go test -component
	docker-compose  -f docker-compose-services.yml -f docker-compose.yml down

.PHONY: docker-test
docker-test:
	docker-compose  -f docker-compose-services.yml -f docker-compose.yml down
	docker build -f Dockerfile . -t template_test --target=test
	docker-compose  -f docker-compose-services.yml -f docker-compose.yml up -d
	docker-compose  -f docker-compose-services.yml -f docker-compose.yml exec -T http go test -v ./...
	docker-compose  -f docker-compose-services.yml -f docker-compose.yml down

.PHONY: docker-local
docker-local:
	docker-compose -f docker-compose-services.yml -f docker-compose-local.yml down
	docker-compose -f docker-compose-services.yml -f docker-compose-local.yml up -d
	docker-compose -f docker-compose-services.yml -f docker-compose-local.yml exec upload-service bash

.PHONY: generate-swagger
generate-swagger:
	swag i -g service/service.go

.PHONY: lint
lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.62.1
	golangci-lint run ./... --timeout 3m --tests=false --skip-dirs=features