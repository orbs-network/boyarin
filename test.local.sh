#!/usr/bin/env bash

docker rm -f $(docker ps -aq)

export DOCKER_IMAGE=${DOCKER_IMAGE-orbs}
export DOCKER_TAG=${DOCKER_TAG-export}

export PEERS=$IP:4400,$IP:4401,$IP:4402
export PEER_KEYS=dfc06c5be24a67adee80b35ab4f147bb1a35c55ff85eda69f40ef827bddec173,92d469d7c004cc0b24a192d9457836bf38effa27536627ef60718b00b0f33152,a899b318e65915aa2de02841eeb72fe51fddad96014b73800ca788a547f8cce0

go run main.go \
    --docker-image $DOCKER_IMAGE \
    --docker-tag $DOCKER_TAG \
    --prefix node1 \
    --config ./e2e-config/node1.json \
    --http-port 8080 \
    --gossip-port 4400 \
    --peers $PEERS \
    --peerKeys $PEER_KEYS

go run main.go \
    --docker-image $DOCKER_IMAGE \
    --docker-tag $DOCKER_TAG \
    --prefix node2 \
    --config ./e2e-config/node2.json \
    --http-port 8081 \
    --gossip-port 4401 \
    --peers $PEERS \
    --peerKeys $PEER_KEYS

go run main.go \
    --docker-image $DOCKER_IMAGE \
    --docker-tag $DOCKER_TAG \
    --prefix node3 \
    --config ./e2e-config/node3.json \
    --http-port 8082 \
    --gossip-port 4402 \
    --peers $PEERS \
    --peerKeys $PEER_KEYS
