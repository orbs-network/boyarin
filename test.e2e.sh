#!/bin/bash -x

docker rm -f $(docker ps -aq)
docker service rm $(docker service ls -q)
docker secret rm $(docker secret ls -q)

rm -rf _tmp

E2E_CONFIG=./e2e-config/ ./e2e.test -test.v
