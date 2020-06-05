#!/bin/bash -e

GITHASH=$(./.circleci/hash.sh)

echo "Building Boyar for git commit ${GITHASH}.."
./build-binaries.sh

echo "Uploading to the dev bucket.."
aws s3 cp --acl public-read boyar.bin s3://boyar-dev-releases/boyar/boyar-$GITHASH.bin