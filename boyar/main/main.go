package main

import (
	"flag"
	"fmt"
	"github.com/orbs-network/boyarin/boyar"
	"github.com/orbs-network/boyarin/supervized"
	"os"
	"time"
)

func main() {
	configUrlPtr := flag.String("config-url", "", "http://my-config/config.json")
	keyPairConfigPathPtr := flag.String("keys", "", "path to public/private key pair in json format")

	daemonizePtr := flag.Bool("daemonize", false, "do not exit the program and keep polling for changes")
	pollingIntervalPtr := flag.Uint("polling-interval", 60, "how often to poll for configuration in daemon mode (in seconds)")

	ethereumEndpointPtr := flag.String("ethereum-endpoint", "", "Ethereum endpoint")
	topologyContractAddressPtr := flag.String("topology-contract-address", "", "Ethereum address for topology contract")

	flag.Parse()

	execute(*daemonizePtr, *keyPairConfigPathPtr, *configUrlPtr, *pollingIntervalPtr, *ethereumEndpointPtr, *topologyContractAddressPtr)
}

func execute(daemonize bool, keyPairConfigPath string, configUrl string, pollingInterval uint, ethereumEndpoint string, topologyContractAddress string) {
	// Even if something crashed, things still were provisioned, meaning the cache should stay
	configCache := make(boyar.BoyarConfigCache)

	if daemonize {
		<-supervized.GoForever(func() {
			if err := boyar.RunOnce(keyPairConfigPath, configUrl, ethereumEndpoint, topologyContractAddress, configCache); err != nil {
				fmt.Println(time.Now(), "ERROR:", err)
			}

			for vcid, hash := range configCache {
				fmt.Println(time.Now(), fmt.Sprintf("Latest successful configuration for vchain %s: %s", vcid, hash))
			}

			<-time.After(time.Duration(pollingInterval) * time.Second)
		})
	} else {
		if err := boyar.RunOnce(keyPairConfigPath, configUrl, ethereumEndpoint, topologyContractAddress, configCache); err != nil {
			fmt.Println(time.Now(), "ERROR:", err)
			os.Exit(1)
		}
	}
}
