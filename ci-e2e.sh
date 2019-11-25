#!/bin/bash -x

GO_VERSION=1.12.9

sudo rm -rf $(dirname $(dirname $(which go))) # remove go base dir (/usr/local/go)  to avoid src duplications
curl -sSL "https://dl.google.com/go/go$GO_VERSION.linux-amd64.tar.gz" | sudo tar -xz -C /usr/local/
curl -sSL "https://github.com/gotestyourself/gotestsum/releases/download/v0.4.0/gotestsum_0.4.0_linux_amd64.tar.gz" | sudo tar -xz -C /usr/local/bin

PATH=$PATH:/usr/local/go/bin # echo "export PATH=$PATH:/usr/local/go/bin" >> $BASH_ENV

go version
env
go env

$(aws ecr get-login --no-include-email --region us-west-2)
docker pull $ORBS_NODE_DOCKER_IMAGE:master
docker pull $SIGNER_DOCKER_IMAGE:master
docker tag $ORBS_NODE_DOCKER_IMAGE:master orbs:export
docker tag $SIGNER_DOCKER_IMAGE:master orbs:signer
docker swarm init
docker rm -f $(docker ps -aq)
docker service rm $(docker service ls -q)
docker secret rm $(docker secret ls -q)

GANACHE_ID=$(docker run -p "7545:7545" -d trufflesuite/ganache-cli -m 'pet talent sugar must audit chief biology trash change wheat educate bone' -i 5777 -p 7545)
#
#export ENABLE_ETHEREUM='true'
#export ETHEREUM_PRIVATE_KEY='7a16631b19e5a7d121f13c3ece279c10c996ff14d8bebe609bf1eca41211b291'
#export ETHEREUM_ENDPOINT='http://ganache:7545'
#export ENABLE_SWARM='true'
#export LOCAL_IP=$(python -c "import socket; print socket.gethostbyname(socket.gethostname())")

# sh sleep 5
# rm -rf _tmp
#gotestsum --junitfile unit-tests.xml
#
#go test ./... -v
#RESULT=$?
#
#docker rm -fv $GANACHE_ID
#exit $RESULT