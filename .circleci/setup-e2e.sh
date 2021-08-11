#!/bin/bash -x

docker pull orbsnetwork/node:experimental
docker pull orbsnetwork/signer:experimental
docker pull orbsnetworkstaging/management-service:v100.0.0

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

# mock EFS

sudo mkdir -p /var/efs
sudo chmod 0777 /var/efs
