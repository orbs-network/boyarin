#!/bin/bash -x

# setup a test machine (written for circleci) for e2e tests
GO_VERSION=1.12.9
NODE_VERSION=12.13

# if go installed
if [ $(hash go 2>/dev/null) ]; then
  echo "go already installed. deleting..."
  sudo rm -rf $(dirname $(dirname $(which go))) # remove go base dir (/usr/local/go) to avoid src duplications
else # go not installed
  echo "go not installed. adding 'usr/local/go/bin' to path..."
  # ensure /usr/local/go/bin is in path (in this step and the next)
  echo "export PATH=$PATH:/usr/local/go/bin" >> $BASH_ENV
  PATH=$PATH:/usr/local/go/bin
fi

# install go and gotestsum
curl -sSL "https://dl.google.com/go/go$GO_VERSION.linux-amd64.tar.gz" | sudo tar -xz -C /usr/local/
curl -sSL "https://github.com/gotestyourself/gotestsum/releases/download/v0.4.0/gotestsum_0.4.0_linux_amd64.tar.gz" | sudo tar -xz -C /usr/local/bin

# if nvm not installed
if [ ! $(hash nvm 2>/dev/null) ]; then
  echo "nvm not installed. installing..."
  $(curl -sSL "https://raw.githubusercontent.com/creationix/nvm/v0.35.1/install.sh") # install nvm
fi
echo "\\nexport NVM_DIR=\"${NVM_DIR}\"\\n[ -s \"\$NVM_DIR/nvm.sh\" ] && \\. \"\$NVM_DIR/nvm.sh\"  # This loads nvm\\n" >> $BASH_ENV # ensure nvm is loaded in the next step
# install node
# https://www.cloudesire.com/how-to-upgrade-node-on-circleci-machine-executor/
if [ -s "$NVM_DIR/nvm.sh" ]; then
  \. "$NVM_DIR/nvm.sh"
fi

nvm install v${NODE_VERSION}
nvm use v${NODE_VERSION}
nvm alias default v${NODE_VERSION}
