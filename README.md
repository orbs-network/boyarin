# Boyar + Strelets

[![CircleCI](https://circleci.com/gh/orbs-network/boyarin/tree/master.svg?style=svg)](https://circleci.com/gh/orbs-network/boyarin/tree/master)

![Boyars, Russian 17th century administrators and warlords](boyars.jpg)

Management layer that provisions virtual chains for [ORBS blockchain](https://github.com/orbs-network/orbs-network-go/).

Works together with [Nebula](https://github.com/orbs-network/nebula).

## Tips

To remove all containers: `docker rm -f $(docker ps -aq)`

## Testing

`./build-binaries.sh && ./test.e2e.sh`
