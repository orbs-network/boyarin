# Boyar + Strelets

[![CircleCI](https://circleci.com/gh/orbs-network/boyarin/tree/master.svg?style=svg)](https://circleci.com/gh/orbs-network/boyarin/tree/master)

![Boyars, Russian 17th century administrators and warlords](boyars.jpg)

Management layer that provisions virtual chains for [ORBS blockchain](https://github.com/orbs-network/orbs-network-go/).

Works together with [Nebula](https://github.com/orbs-network/nebula).

## Tips

To remove all containers: `docker rm -f $(docker ps -aq)`

## Testing

`./build-binaries.sh && ./test.e2e.sh`

## CLI options

`--config-url` path to Boyar configuration

`--daemonize` do not exit the program and keep polling for changes

`--ethereum-endpoint` HTTP endpoint for the Ethereum node

`--topology-contract-address` Ethereum address for topology contract

`--keys` path to address/private key pair in json format (example in `e2e-config/node1/keys.json`)

`--polling-interval` how often to poll for configuration in daemon mode (in seconds) (default 60)

`--orchestrator-options` allows to override `orchestrator-options` section of boyar JSON config. Takes JSON object as a parameter.

`--show-configuration` Show configuration for evaluation and exit

### SSL options

`--ssl-certificate` path to SSL certificate

`--ssl-private-key` path to SSL private key

If both these parameters are present, the node will also start service SSL traffic.

### Running as a daemon

    boyar --config-url https://s3.amazonaws.com/boyar-bootstrap-test/boyar/config.json \
        --keys ./e2e-config/node3/keys.json \
        --daemonize

### Running as a daemon and fetching topology from Ethereum

    boyar --config-url https://s3.amazonaws.com/boyar-bootstrap-test/boyar/config.json \
        --keys ./e2e-config/node3/keys.json \
        --ethereum-endpoint http://localhost:7545 \
        --topology-contract-address 0x409aa7d40dfcfa3725d722a720ff1ba147df4bec \
        --daemonize

### Print configuration and exit

    boyar --config-url https://s3.amazonaws.com/boyar-bootstrap-test/boyar/config.json \
        --keys ./e2e-config/node3/keys.json \
        --ethereum-endpoint http://localhost:7545 \
        --topology-contract-address 0x409aa7d40dfcfa3725d722a720ff1ba147df4bec \
        --show-configuration