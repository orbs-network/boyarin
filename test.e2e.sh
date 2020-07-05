#!/bin/bash

cleanup(){
    echo "cleaning up child processes"
    kill $(pgrep -P $$) # kill all processes whose parent PID is the current PID
    exit
}

# cleanup on ctr+C
trap cleanup SIGINT # will execute the function cleanup when CTRL-C is pressed

# assert/prepare this machine to run the test
GO_VERSION=1.12.9
NODE_VERSION=12.13

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


if ! [ -x "$(command -v gotestsum)" ]; then
  echo "gotestsum not installed. Installing"
  GO111MODULE=off go get gotest.tools/gotestsum
fi

.circleci/setup-e2e.sh

export ENABLE_SWARM=true

gotestsum ./... -- -p 1 &
#go test -p 1 ./... &

wait $! # wait for the last command launched in background

cleanup

