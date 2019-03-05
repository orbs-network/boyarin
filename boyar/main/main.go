package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/orbs-network/boyarin/boyar"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/supervized"
	"os"
	"time"
)

type flags struct {
	configUrl         string
	keyPairConfigPath string

	daemonize bool

	pollingInterval    time.Duration
	timeout            time.Duration
	maxReloadTimeDelay time.Duration

	ethereumEndpoint        string
	topologyContractAddress string
}

func main() {
	configUrlPtr := flag.String("config-url", "", "http://my-config/config.json")
	keyPairConfigPathPtr := flag.String("keys", "", "path to public/private key pair in json format")

	daemonizePtr := flag.Bool("daemonize", false, "do not exit the program and keep polling for changes")
	pollingIntervalPtr := flag.Duration("polling-interval", 1*time.Minute, "how often to poll for configuration in daemon mode (duration: 1s, 1m, 1h, etc)")
	maxReloadTimePtr := flag.Duration("max-reload-time-delay", 15*time.Minute, "introduces jitter to reloading configuration to make network more stable, only works in daemon mode (duration: 1s, 1m, 1h, etc)")

	timeoutPtr := flag.Duration("timeout", 10*time.Minute, "timeout for provisioning all virtual chains (duration: 1s, 1m, 1h, etc)")

	ethereumEndpointPtr := flag.String("ethereum-endpoint", "", "Ethereum endpoint")
	topologyContractAddressPtr := flag.String("topology-contract-address", "", "Ethereum address for topology contract")

	showConfiguration := flag.Bool("show-configuration", false, "Show configuration and exit")
	help := flag.Bool("help", false, "Show usage")

	flag.Parse()

	flags := &flags{
		configUrl:               *configUrlPtr,
		keyPairConfigPath:       *keyPairConfigPathPtr,
		daemonize:               *daemonizePtr,
		pollingInterval:         *pollingIntervalPtr,
		timeout:                 *timeoutPtr,
		maxReloadTimeDelay:      *maxReloadTimePtr,
		ethereumEndpoint:        *ethereumEndpointPtr,
		topologyContractAddress: *topologyContractAddressPtr,
	}

	if *help {
		flag.Usage()
		return
	}

	if *showConfiguration {
		printConfiguration(flags)
		return
	}

	execute(flags)
}

func printConfiguration(flags *flags) {
	cfg, err := config.GetConfiguration(flags.configUrl, flags.ethereumEndpoint, flags.topologyContractAddress, "")
	if err != nil {
		fmt.Println("ERROR: could not pull valid configuration:", err)
		return
	}

	fmt.Println("# Orchestrator options:\n# ============================")
	orchestratorOptions, _ := json.MarshalIndent(cfg.OrchestratorOptions(), "", "  ")
	fmt.Println(string(orchestratorOptions))

	fmt.Println("# Peers:\n# ============================")
	peers, _ := json.MarshalIndent(cfg.FederationNodes(), "", "  ")
	fmt.Println(string(peers))

	fmt.Println("# Chains:\n# ============================")
	chains, _ := json.MarshalIndent(cfg.Chains(), "", "  ")
	fmt.Println(string(chains))
}

func execute(flags *flags) {
	if flags.configUrl == "" {
		fmt.Println("--config-url is a required parameter for provisioning flow")
		os.Exit(1)
	}

	if flags.keyPairConfigPath == "" {
		fmt.Println("--keys is a required parameter for provisioning flow")
		os.Exit(1)
	}

	// Even if something crashed, things still were provisioned, meaning the cache should stay
	configCache := make(config.BoyarConfigCache)

	if flags.daemonize {
		<-supervized.GoForever(func() {
			for first := true; ; first = false {
				cfg, err := config.GetConfiguration(flags.configUrl, flags.ethereumEndpoint, flags.topologyContractAddress, flags.keyPairConfigPath)
				if err != nil {
					fmt.Println(time.Now(), "ERROR:", fmt.Errorf("could not generate configuration: %s", err))
				} else {
					// skip delay when provisioning for the first time when the node goes up
					if !first {
						reloadTimeDelay := cfg.ReloadTimeDelay(flags.maxReloadTimeDelay)
						fmt.Println(fmt.Sprintf("INFO: waiting for %s to apply new configuration", reloadTimeDelay))
						<-time.After(reloadTimeDelay)
					}

					ctx, cancel := context.WithTimeout(context.Background(), flags.timeout)
					defer cancel()

					if err := boyar.FullFlow(ctx, cfg, configCache); err != nil {
						fmt.Println(time.Now(), "ERROR:", err)
					}

					for vcid, hash := range configCache {
						fmt.Println(time.Now(), fmt.Sprintf("Latest successful configuration for vchain %s: %s", vcid, hash))
					}
				}

				<-time.After(flags.pollingInterval)
			}
		})
	} else {
		cfg, err := config.GetConfiguration(flags.configUrl, flags.ethereumEndpoint, flags.topologyContractAddress, flags.keyPairConfigPath)
		if err != nil {
			fmt.Println(time.Now(), "ERROR:", fmt.Errorf("could not generate configuration: %s", err))
			os.Exit(1)
		}

		ctx, cancel := context.WithTimeout(context.Background(), flags.timeout)
		defer cancel()

		if err := boyar.FullFlow(ctx, cfg, configCache); err != nil {
			fmt.Println(time.Now(), "ERROR:", err)
			os.Exit(1)
		}
	}
}
