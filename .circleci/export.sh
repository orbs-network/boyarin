#!/bin/bash -xe

export VERSION=$(cat .version)
export BOYAR_S3_PATH=s3://orbs-network-releases/infrastructure/boyar/boyar-$VERSION.bin

aws s3 cp --acl public-read ./_bin/boyar-$VERSION.bin $BOYAR_S3_PATH
