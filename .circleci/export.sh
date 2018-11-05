#!/bin/bash -xe

export S3_PATH=s3://orbs-network-releases/infrastructure/strelets/strelets-$(git rev-parse HEAD).bin

aws s3 cp --acl public-read ./strelets.bin $S3_PATH
