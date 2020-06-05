#!/bin/sh -xe

export SEMVER=$(cat ./.version)
export CONFIG_PKG="github.com/orbs-network/boyarin/version"

go build -ldflags "-w -extldflags '-static' -X $CONFIG_PKG.SemanticVersion=$SEMVER -X $CONFIG_PKG.CommitVersion=$GIT_COMMIT" -o boyar.bin -a ./boyar/main/main.go
