#!/bin/bash

# setup a test machine (written for circleci) for e2e tests
GO_VERSION=1.12.9
NODE_VERSION=12.13

trap terminate SIGINT # will execute the function terminate when CTRL-C is pressed
terminate(){
    echo "cleaning up child processes"
    kill $(pgrep -P $$) # kill all processes whose parent PID is the current PID
    exit
}

if ! [ -x "$(command -v go)" ]; then
  echo "go not installed"
  exit 1
fi

if ! [ -x "$(command -v docker)" ]; then
  echo "docker not installed"
  exit 1
fi

if (! docker stats --no-stream ); then
  echo "docker not running"
  exit 1
fi

if ! [ -x "$(command -v aws)" ]; then
  echo "aws cli not installed"
  exit 1
fi

if ! [ -x "$(command -v ganache-cli)" ]; then
  echo "ganache-cli not installed"
  if ! [ -x "$(command -v npm)" ]; then
    echo "npm not installed"
    exit 1
  fi
  echo "instaling ganache-cli"
  npm install -g ganache-cli
fi

if ! [ -x "$(command -v gotestsum)" ]; then
  echo "gotestsum not installed. Installing"
  GO111MODULE=off go get gotest.tools/gotestsum
fi

nc -z 127.0.0.1 7545
if [ "$?" -eq 0 ] ; then
  echo "port 7545 (needed for ganache-cli) is open in local machine"
  exit 1
fi

ganache-cli -m 'pet talent sugar must audit chief biology trash change wheat educate bone' -h 0.0.0.0  -i 5777 -p 7545 & # run ganache in the background

./setup-e2e.sh

export ENABLE_SWARM=true
export ENABLE_ETHEREUM=true
export ETHEREUM_PRIVATE_KEY=7a16631b19e5a7d121f13c3ece279c10c996ff14d8bebe609bf1eca41211b291
export ETHEREUM_ENDPOINT=http://localhost:7545
# export CI=true # some tests are skipped in CI

gotestsum ./... -- -p 1 &
#go test -p 1 ./... &

wait $! # wait for the last command launched in background

terminate # cleanup

