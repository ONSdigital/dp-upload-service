#!/bin/bash -eux

pushd dp-upload-service
  make build
  cp build/dp-upload-service Dockerfile.concourse ../build
popd
