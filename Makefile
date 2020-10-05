BINPATH ?= build

BUILD_TIME=$(shell date +%s)
GIT_COMMIT=$(shell git rev-parse HEAD)
VERSION ?= $(shell git tag --points-at HEAD | grep ^v | head -n 1)

LDFLAGS = -ldflags "-X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -X main.Version=$(VERSION)"

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

.PHONY: convey
convey:
	goconvey ./...

