#!/bin/sh -xe

rm -rf _tmp

go test ./boyar/topology/... -v
go test ./test/e2e/... -v
