#!/bin/bash -x

# setup a test machine (written for circleci) for e2e tests
GO_VERSION=1.12.9

# install go and gotestsum
sudo rm -rf $(dirname $(dirname $(which go))) # remove go base dir (/usr/local/go) to avoid src duplications
curl -sSL "https://dl.google.com/go/go$GO_VERSION.linux-amd64.tar.gz" | sudo tar -xz -C /usr/local/
curl -sSL "https://github.com/gotestyourself/gotestsum/releases/download/v0.4.0/gotestsum_0.4.0_linux_amd64.tar.gz" | sudo tar -xz -C /usr/local/bin

# ensure /usr/local/go/bin is in path (in this step and the next)
echo "export PATH=$PATH:/usr/local/go/bin" >> $BASH_ENV
PATH=$PATH:/usr/local/go/bin
