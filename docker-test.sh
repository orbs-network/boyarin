#!/bin/bash -x

docker rm -f $(docker ps -aq)
docker service rm $(docker service ls -q)
docker secret rm $(docker secret ls -q)

export LOCAL_IP=$(python -c "import socket; print socket.gethostbyname(socket.gethostname())")

docker-compose -f docker-compose.test.yml up --abort-on-container-exit --exit-code-from boyar
