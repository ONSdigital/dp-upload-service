#!/bin/bash -eux

pushd dp-upload-service
  go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.43.0
  make lint
popd