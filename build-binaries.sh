#!/bin/sh -xe

export GIT_COMMIT=$(git rev-parse HEAD)
export SEMVER=$(cat ./.version)
export CONFIG_PKG="github.com/orbs-network/boyarin/boyar/config"

time go build -ldflags '-w -extldflags "-static"' -ldflags "-X $CONFIG_PKG.SemanticVersion=$SEMVER -X $CONFIG_PKG.CommitVersion=$GIT_COMMIT" -o boyar.bin -a ./boyar/main/main.go
