#!/bin/bash -x

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
