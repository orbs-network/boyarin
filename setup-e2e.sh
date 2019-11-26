#!/bin/bash -x

# install go and gotestsum (only in CI)
GO_VERSION=1.12.9
if [ "${CI:-false}" = "true" ]; then
  sudo rm -rf $(dirname $(dirname $(which go))) # remove go base dir (/usr/local/go)  to avoid src duplications
  curl -sSL "https://dl.google.com/go/go$GO_VERSION.linux-amd64.tar.gz" | sudo tar -xz -C /usr/local/
  curl -sSL "https://github.com/gotestyourself/gotestsum/releases/download/v0.4.0/gotestsum_0.4.0_linux_amd64.tar.gz" | sudo tar -xz -C /usr/local/bin

  # ensure /usr/local/go/bin is in path (in this step and the next)
  echo "export PATH=$PATH:/usr/local/go/bin" >> $BASH_ENV
  PATH=$PATH:/usr/local/go/bin
fi

# print environment to log
go version
env
go env

# setup
$(aws ecr get-login --no-include-email --region us-west-2)
docker pull ${ORBS_NODE_DOCKER_IMAGE}:master
docker pull ${SIGNER_DOCKER_IMAGE}:master
docker tag ${ORBS_NODE_DOCKER_IMAGE}:master orbs:export
docker tag ${SIGNER_DOCKER_IMAGE}:master orbs:signer
docker swarm init

# clean docker state
DOCKER_INSTANCES=$(docker ps -aq)
DOCKER_SERVICES=$(docker service ls -q)
DOCKER_SECRETS=$(docker secret ls -q)
if [ -n "${DOCKER_INSTANCES}" ]; then
  docker rm -f ${DOCKER_INSTANCES}
fi
if [ -n "${DOCKER_SERVICES}" ]; then
  docker service rm ${DOCKER_SERVICES}
fi
if [ -n "${DOCKER_SECRETS}" ]; then
  docker secret rm ${DOCKER_SECRETS}
fi
