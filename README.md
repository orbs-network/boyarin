# Boyar + Strelets

[![CircleCI](https://circleci.com/gh/orbs-network/boyarin/tree/master.svg?style=svg)](https://circleci.com/gh/orbs-network/boyarin/tree/master)

![Boyars, Russian 17th century administrators and warlords](boyars.jpg)

Management layer that provisions virtual chains for [ORBS blockchain](https://github.com/orbs-network/orbs-network-go/). Works together with [Polygon](https://github.com/orbs-network/polygon).

Architecture and workflows implemented are described in relevant parts of [ORBS spec](https://github.com/orbs-network/orbs-spec/blob/master/node-architecture/BOYAR.md).

## Changelog

### Breaking changes in v1.10.0

`orchestrator` is now **mandatory** part of the config. If this section is not found in the config, Boyar will treat it as invalid and can revert to bootstrap flow (if `boostrap-reset-timeout` option was used).

### Breaking changes in v1.0.0

Services has become **mandatory** part of the config.

### v0.17.0

Staring from version 0.17.0, Boyar only works with Docker version higher than 19.03.

## Building

Building in Docker:

```
./docker-build.sh
```

Alternative faster build:

```bash
export GOOS=linux
./build-binaries.sh
```

## Tips

To remove all containers: `docker rm -f $(docker ps -aq)`

## Testing

`./build-binaries.sh && ./test.e2e.sh`

In case you ever need to regenerate the SSL certificate:

`openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 10000 -nodes`

## CLI options

`--log` path to log file, otherwise will log to stdout

`--status` path to status file

`--metrics` path to metrics file

`--config-url` path to Boyar configuration

`--ethereum-endpoint` HTTP endpoint for the Ethereum node

`--topology-contract-address` legacy parameter, will be removed later

`--keys` path to address/private key pair in json format (example in `e2e-config/node1/keys.json`)

`--polling-interval` how often to poll for configuration in daemon mode (in seconds) (default 60)

`--orchestrator-options` allows to override `orchestrator` section of boyar JSON config. Takes JSON object as a parameter.

`--show-configuration` show configuration for evaluation and exit

`--show-status` print status in json format and exit

`--max-reload-time-delay` introduces jitter to reloading configuration to make network more stable, only works in daemon mode (duration: 1s, 1m, 1h, etc)

`--timeout` timeout for provisioning all virtual chains (duration: 1s, 1m, 1h, etc)

`--auto-update` enables boyar binary auto update (default false)

`--shutdown-after-update` the process shuts down after automatic update is performed and **DOES NOT** restart; recommended to be used with an external process manager (default false)

`--bootstrap-reset-timeout` if the process is unable to receive valid configuration within a limited timeframe (duration: 1s, 1m, 1h, etc), it will exit with an error; recommended to be used with an external process manager (default: 30m)

`--version` show version, git commit and Docker API version

### SSL options

`--ssl-certificate` path to SSL certificate

`--ssl-private-key` path to SSL private key

If both these parameters are present, the node will also start service SSL traffic.

### Running as a daemon

    boyar --config-url https://s3.amazonaws.com/boyar-bootstrap-test/boyar/config.json \
        --keys ./e2e-config/node3/keys.json \
        --daemonize

It is recommended to run Boyar together with some kind of process manager (for example, [Supervisord](http://supervisord.org)).
If autoupdate is enabled, it becomes crucial if you enable `--shutdown-after-update` feature for seamless automatic updates.

### Print configuration and exit

    boyar --config-url https://s3.amazonaws.com/boyar-bootstrap-test/boyar/config.json \
        --keys ./e2e-config/node3/keys.json \
        --ethereum-endpoint http://localhost:7545 \
        --topology-contract-address 0x409aa7d40dfcfa3725d722a720ff1ba147df4bec \
        --show-configuration

## Boyar config

```json
{
  "network": [ // network topology, usually taken from Ethereum
    {
      "address":"dfc06c5be24a67adee80b35ab4f147bb1a35c55ff85eda69f40ef827bddec173",
      "ip":"192.168.1.14"
    }
  ],
  "orchestrator": { // orchestrator options (right now only Docker Swarm is supported)
    "storage-driver": "local", // storage driver for docker
    "storage-mount-type": "bind", // mounts to /var/efs
    "storage-options": { // parameters passed to storage driver (optional)
      "maxRetries": "10"
    },
    "max-reload-time-delay": "1m", // optional
    "ExecutableImage": { // optional
      "Url": "https://github.com/orbs-network/boyarin/releases/download/v1.8.0/boyar-v1.8.0.bin",
      "Sha256": "0d7df92307b95ff7e2923dd7509e3b5bac23deb491b5c08d522b11ac08d78e02"
    }
  },
  "chains": [
    {
      "Id":         42, // vchain id passed to the binary inside the container (mandatory, unique)
      "InternalPort":   4400, // gossip port passed to the binary inside the container (mandatory, unique)
      "ExternalPort": 4400, // gossip port passed to the binary inside the container (mandatory, unique)
      "Disabled": false, // (optional)
      "PurgeData": false, // destroys all data related to the chain (logs, cache, status, blocks), only works with EFS (optional)
      "DockerConfig": {
        "ContainerNamePrefix": "orbs-network",
        "Image":  "orbsnetwork/node", // Docker image
        "Tag":    "v1.1.0", // Docker tag
        "Pull":   true, // Pull new Docker image during provisioning
        "Resources": { // Docker limits (optional)
          "Limits": { // maximum available values (optional)
            "Memory": 1024, // in Mb
            "CPUs": 1 // in shares, 1 being 100% of a single CPU
          },
          "Reservations": { // reserved resources (optional)
            "Memory": 512,
            "CPUs": 0.5
          }
        },
        "Volumes": { // volume size settings (optional)
          "Blocks": 5, // in Gb
          "Logs": 1 // in Gb
        }
      },
      "Config": { // configuration passed to the binary inside the container
        "active-consensus-algo": 2
      }
    }
  ],
  "services": { // list of auxilary services (mandatory)
    "signer": {
      "Port": 7777,
      "DockerConfig": {
        "ContainerNamePrefix": "signer",
        "Image":  "orbsnetwork/orbs-network-signer",
        "Tag":    "v1.1.0",
        "Pull":   true,
        "Resources": {
          "Limits": {
            "Memory": 1024,
            "CPUs": 1
          },
          "Reservations": {
            "Memory": 512,
            "CPUs": 0.5
          }
        }
      },
      "Config": { // configuration passed to the binary inside the container
        "api": "v1"
      }
    },
    "service-name": {
      "InternalPort": 8080,
      "ExternalPort": 2000,
      "InjectNodePrivateKey": false, // should pass private key as a file; **never** set it to true, default false (optional)
      "ExecutablePath": "/opt/orbs/service", // default (optional)
      "AllowAccessToSigner": false, // should be able communicate with the signer service, default false (optional)
      "AllowAccessToServices": true, // should be able to communicate with other services, default true (optional)
      "MountNodeLogs": false, // mounts all service and vchain logs inside the container, default false (optional)
      "Disabled": false, // (optional)
      "PurgeData": false, // destroys all data related to the service (logs, cache, status), only works with EFS (optional)
      "DockerConfig": {
        "Image": "orbsnetwork/service-name",
        "Tag": "latest",
        "Pull": false
      },
      "Config": {
      }
    }
  }
}
```

### Build and Release

#### Prerequisites 
- go1.13 or later must be installed
- Access to this repository

#### Prepare
- .version file should be updated
- commit and push all changes to git. During build we rely on the commit hash.
- These instructions are for target env of linux/amd64. For other os/arch, set `GOOS` and `GOARCH` accordingly 

#### Build 
```bash
GOARCH=amd64 GOOS=linux ./build-binaries.sh
```

#### Release 
- Create a new release in this repository
- Attach all files found in `_bin` folder as attachments to the new release
- In order to deploy the new version on existing and new Orbs nodes, a new version of [Management Service](https://github.com/orbs-network/management-service) must be deployed referencing the new versions binary and checksum files. see [here](https://github.com/orbs-network/management-service) for more details.
