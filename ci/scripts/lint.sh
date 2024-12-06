#!/bin/bash -eux

pushd dp-upload-service
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.62.1
  make lint
popd