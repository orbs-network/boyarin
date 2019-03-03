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

	pollingInterval time.Duration
	timeout         time.Duration

	ethereumEndpoint        string
	topologyContractAddress string
}

func main() {
	flags := &flags{
		configUrl:               *flag.String("config-url", "", "http://my-config/config.json"),
		keyPairConfigPath:       *flag.String("keys", "", "path to public/private key pair in json format"),
		daemonize:               *flag.Bool("daemonize", false, "do not exit the program and keep polling for changes"),
		pollingInterval:         *flag.Duration("polling-interval", 1*time.Minute, "how often to poll for configuration in daemon mode (duration: 1s, 1m, 1h, etc)"),
		timeout:                 *flag.Duration("timeout", 10*time.Minute, "timeout for provisioning all virtual chains (duration: 1s, 1m, 1h, etc)"),
		ethereumEndpoint:        *flag.String("ethereum-endpoint", "", "Ethereum endpoint"),
		topologyContractAddress: *flag.String("topology-contract-address", "", "Ethereum address for topology contract"),
	}

	showConfiguration := flag.Bool("show-configuration", false, "Show configuration and exit")
	help := flag.Bool("help", false, "Show usage")

	flag.Parse()

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
			ctx, cancel := context.WithTimeout(context.Background(), flags.timeout)
			defer cancel()

			cfg, err := config.GetConfiguration(flags.configUrl, flags.ethereumEndpoint, flags.topologyContractAddress, flags.keyPairConfigPath)
			if err != nil {
				fmt.Println(time.Now(), "ERROR:", fmt.Errorf("could not generate configuration: %s", err))
			} else {
				if err := boyar.FullFlow(ctx, cfg, configCache); err != nil {
					fmt.Println(time.Now(), "ERROR:", err)
				}

				for vcid, hash := range configCache {
					fmt.Println(time.Now(), fmt.Sprintf("Latest successful configuration for vchain %s: %s", vcid, hash))
				}
			}

			<-time.After(flags.pollingInterval)
		})
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), flags.timeout)
		defer cancel()

		cfg, err := config.GetConfiguration(flags.configUrl, flags.ethereumEndpoint, flags.topologyContractAddress, flags.keyPairConfigPath)
		if err != nil {
			fmt.Println(time.Now(), "ERROR:", fmt.Errorf("could not generate configuration: %s", err))
			os.Exit(1)
		} else if err := boyar.FullFlow(ctx, cfg, configCache); err != nil {
			fmt.Println(time.Now(), "ERROR:", err)
			os.Exit(1)
		}
	}
}
