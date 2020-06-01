#!/bin/bash -x

SIGNER_DOCKER_IMAGE=${SIGNER_DOCKER_IMAGE:-"727534866935.dkr.ecr.us-west-2.amazonaws.com/orbs-network-signer"}
ORBS_NODE_DOCKER_IMAGE=${ORBS_NODE_DOCKER_IMAGE:-"727534866935.dkr.ecr.us-west-2.amazonaws.com/orbs-network-v1"}

#echo "downloading latest images from AWS ECR"
#$(aws ecr get-login --no-include-email --region us-west-2)
#docker pull ${ORBS_NODE_DOCKER_IMAGE}:experimental
#docker pull ${SIGNER_DOCKER_IMAGE}:experimental
#docker tag ${ORBS_NODE_DOCKER_IMAGE}:experimental orbs:export
#docker tag ${SIGNER_DOCKER_IMAGE}:experimental orbs:signer

docker pull orbsnetwork/node:experimental
docker pull orbsnetwork/signer:experimental
docker pull orbsnetworkstaging/management-service:v100.0.0
docker pull orbsnetwork/ethereum-light-client-service:latest

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

docker container prune -f
docker volume prune -f
docker network prune -f