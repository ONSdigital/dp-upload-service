#!/bin/bash -eux

pushd dp-upload-service
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.43.0
  make lint
popd