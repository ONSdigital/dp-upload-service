#!/bin/bash -eux

export cwd=$(pwd)

pushd $cwd/dp-upload-service
  make audit
popd 