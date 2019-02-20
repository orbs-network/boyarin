#!/bin/bash -xe

export VERSION=$(git rev-parse HEAD)
export BOYAR_S3_PATH=s3://orbs-network-releases/infrastructure/boyar/boyar-$VERSION.bin

aws s3 cp --acl public-read --profile admin ./boyar.bin $BOYAR_S3_PATH