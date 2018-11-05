#!/bin/bash -x

docker rm -f $(docker ps -aq)
rm -rf _tmp

export DOCKER_IMAGE=${DOCKER_IMAGE-orbs}
export DOCKER_TAG=${DOCKER_TAG-export}

export IP=$(python -c "import socket; print socket.gethostbyname(socket.gethostname())")

export PEERS=$IP:4400,$IP:4401,$IP:4402
export PEER_KEYS=dfc06c5be24a67adee80b35ab4f147bb1a35c55ff85eda69f40ef827bddec173,92d469d7c004cc0b24a192d9457836bf38effa27536627ef60718b00b0f33152,a899b318e65915aa2de02841eeb72fe51fddad96014b73800ca788a547f8cce0

./strelets.bin provision-virtual-chain \
    --chain-config ./e2e-config/node1/vchain.json \
    --keys-config ./e2e-config/node1/keys.json \
    --peers $PEERS \
    --peerKeys $PEER_KEYS

./strelets.bin provision-virtual-chain \
    --chain-config ./e2e-config/node2/vchain.json \
    --keys-config ./e2e-config/node2/keys.json \
    --peers $PEERS \
    --peerKeys $PEER_KEYS

./strelets.bin provision-virtual-chain \
    --chain-config ./e2e-config/node3/vchain.json \
    --keys-config ./e2e-config/node3/keys.json \
    --peers $PEERS \
    --peerKeys $PEER_KEYS

python2.7 test/e2e_test.py
