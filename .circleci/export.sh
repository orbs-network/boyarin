#!/bin/bash -xe

export VERSION=$(git rev-parse HEAD)
export STRELETS_S3_PATH=s3://orbs-network-releases/infrastructure/strelets/strelets-$VERSION.bin
export BOYAR_S3_PATH=s3://orbs-network-releases/infrastructure/boyar/boyar-$VERSION.bin

aws s3 cp --acl public-read ./strelets.bin $STRELETS_S3_PATH
aws s3 cp --acl public-read ./boyar.bin $BOYAR_S3_PATH