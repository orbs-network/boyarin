#!/bin/bash -x

docker rm -f $(docker ps -aq)
rm -rf _tmp

E2E_CONFIG=./e2e-config/ ./e2e.test -test.v
