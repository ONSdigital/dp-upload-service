#!/bin/bash
set -x
awslocal s3 mb s3://testing
awslocal s3 mb s3://deprecated
set +x