#!/bin/bash -eux

pushd dp-upload-service
  make test-component
popd
