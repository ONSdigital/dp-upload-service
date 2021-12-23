#!/bin/bash -eux

pushd dp-upload-service
  make docker-test-component
popd
