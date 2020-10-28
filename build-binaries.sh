#!/bin/sh -xe

export SEMVER=$(cat ./.version)
export GIT_COMMIT=${GIT_COMMIT-$(./.circleci/hash.sh)}
export CONFIG_PKG="github.com/orbs-network/boyarin/version"
export CGO_ENABLED=0

go build -ldflags "-w -extldflags '-static' -X $CONFIG_PKG.SemanticVersion=$SEMVER -X $CONFIG_PKG.CommitVersion=$GIT_COMMIT" -tags "$BUILD_FLAG usergo netgo" -o boyar.bin -a ./boyar/main/main.go
