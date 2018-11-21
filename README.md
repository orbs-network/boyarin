# Boyar + Strelets

[![CircleCI](https://circleci.com/gh/orbs-network/boyarin/tree/master.svg?style=svg)](https://circleci.com/gh/orbs-network/boyarin/tree/master)

![Boyars, Russian 17th century administrators and warlords](boyars.jpg)

Management layer that provisions virtual chains for [ORBS blockchain](https://github.com/orbs-network/orbs-network-go/).

## CLI

To create new virtual chain

```
strelets provision-virtual-chain \
    --chain 42 \
    --prefix node1 \
    --config ./e2e-config/node1.json \
    --http-port 8080 \
    --gossip-port 4400 \
    --peers 10.4.12.46:4400 \
    --peerKeys dfc06c5be24a67adee80b35ab4f147bb1a35c55ff85eda69f40ef827bddec173 \
    --docker-image orbs \
    --docker-tag export
```

To remove already provisioned virtual chain

```
strelets remove-virtual-chain \
    --chain 42
```

## Tips

To remove all containers: `docker rm -f $(docker ps -aq)`

## Testing

`./build-binaries.sh && ./test.e2e.sh`
