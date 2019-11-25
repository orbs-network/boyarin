#!/bin/bash -x

./docker-build.sh
$(aws ecr get-login --no-include-email --region us-west-2)
docker pull $ORBS_NODE_DOCKER_IMAGE:master
docker pull $SIGNER_DOCKER_IMAGE:master
docker tag $ORBS_NODE_DOCKER_IMAGE:master orbs:export
docker tag $SIGNER_DOCKER_IMAGE:master orbs:signer
docker swarm init
docker rm -f $(docker ps -aq)
docker service rm $(docker service ls -q)
docker secret rm $(docker secret ls -q)

export LOCAL_IP=$(python -c "import socket; print socket.gethostbyname(socket.gethostname())")

docker-compose -f docker-compose.test.yml up --abort-on-container-exit --exit-code-from boyar
