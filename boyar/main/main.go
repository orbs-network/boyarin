package main

import (
	"encoding/json"
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

	showConfiguration := flag.Bool("show-configuration", false, "Show configuration and exit")

	help := flag.Bool("help", false, "Show usage")

	flag.Parse()

	if *help {
		flag.Usage()
		return
	}

	if *showConfiguration {
		printConfiguration(*configUrlPtr, *ethereumEndpointPtr, *topologyContractAddressPtr)
		return
	}

	execute(*daemonizePtr, *keyPairConfigPathPtr, *configUrlPtr, *pollingIntervalPtr, *ethereumEndpointPtr, *topologyContractAddressPtr)
}

func printConfiguration(configUrl string, ethereumEndpoint string, topologyContractAddress string) {
	config, err := boyar.GetConfiguration(configUrl, ethereumEndpoint, topologyContractAddress)
	if err != nil {
		fmt.Println("ERROR: could not pull valid configuration:", err)
		return
	}

	fmt.Println("# Orchestrator options:\n# ============================")
	orchestratorOptions, _ := json.MarshalIndent(config.OrchestratorOptions(), "", "  ")
	fmt.Println(string(orchestratorOptions))

	fmt.Println("# Peers:\n# ============================")
	peers, _ := json.MarshalIndent(config.FederationNodes(), "", "  ")
	fmt.Println(string(peers))

	fmt.Println("# Chains:\n# ============================")
	chains, _ := json.MarshalIndent(config.Chains(), "", "  ")
	fmt.Println(string(chains))
}

func execute(daemonize bool, keyPairConfigPath string, configUrl string, pollingInterval uint, ethereumEndpoint string, topologyContractAddress string) {
	if configUrl == "" {
		fmt.Println("--config-url is a required parameter for provisioning flow")
		os.Exit(1)
	}

	if keyPairConfigPath == "" {
		fmt.Println("--keys is a required parameter for provisioning flow")
		os.Exit(1)
	}

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
