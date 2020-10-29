#!/bin/bash -xe

docker build --build-arg git_commit=$(./.circleci/hash.sh) -f Dockerfile.build -t orbs:build .

[ "$(docker ps -a | grep orbs_build)" ] && docker rm -f orbs_build

docker run --name orbs_build orbs:build sleep 1

export SRC=/src

rm -rf _bin
docker cp orbs_build:$SRC/_bin .

