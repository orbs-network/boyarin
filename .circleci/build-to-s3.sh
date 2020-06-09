#!/bin/bash -xe

GITHASH=$(./.circleci/hash.sh)

export BUILD_FLAG="$BUILD_FLAG netgo osusergo" # allows static linking, further reading https://github.com/golang/go/issues/30419

echo "Building Boyar for git commit ${GITHASH}.."
./build-binaries.sh

echo "Uploading to the dev bucket.."
aws s3 cp --acl public-read boyar.bin s3://boyar-dev-releases/boyar/boyar-$GITHASH.bin

echo "Built and uploaded!"