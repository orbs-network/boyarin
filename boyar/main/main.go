package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/orbs-network/boyarin/boyar"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/supervized"
	"github.com/orbs-network/scribe/log"
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

	loggerHttpEndpoint string
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

	loggerHttpEndpointPtr := flag.String("logger-http-endpoint", "", "")

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
		loggerHttpEndpoint:      *loggerHttpEndpointPtr,
	}

	logger, err := getLogger(flags)
	if err != nil {
		os.Exit(1)
	}

	if *help {
		flag.Usage()
		return
	}

	if *showConfiguration {
		printConfiguration(flags, logger)
		return
	}

	if err := execute(flags, logger); err != nil {
		logger.Error("Startup failure", log.Error(err))
		os.Exit(1)
	}
}

func getLogger(flags *flags) (log.Logger, error) {
	outputs := []log.Output{log.NewFormattingOutput(os.Stdout, log.NewHumanReadableFormatter())}

	if flags.loggerHttpEndpoint != "" {
		outputs = append(outputs, log.NewBulkOutput(
			log.NewHttpWriter(flags.loggerHttpEndpoint),
			log.NewJsonFormatter().WithTimestampColumn("@timestamp"), 1))
	}

	logger := log.GetLogger().
		WithTags(log.String("app", "boyar")).
		WithOutput(outputs...).
		WithSourcePrefix("boyarin/")

	cfg, _ := config.NewStringConfigurationSource("{}", "")
	cfg.SetKeyConfigPath(flags.keyPairConfigPath)
	if err := cfg.VerifyConfig(); err != nil {
		logger.Error("Invalid configuration", log.Error(err))
		return nil, err
	}

	return logger.WithTags(log.Node(string(cfg.NodeAddress()))), nil
}

func printConfiguration(flags *flags, logger log.Logger) {
	cfg, err := config.GetConfiguration(flags.configUrl, flags.ethereumEndpoint, flags.topologyContractAddress, flags.keyPairConfigPath)
	if err != nil {
		logger.Error("could not pull valid configuration", log.Error(err))
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

func execute(flags *flags, logger log.Logger) error {
	if flags.configUrl == "" {
		return fmt.Errorf("--config-url is a required parameter for provisioning flow")
	}

	if flags.keyPairConfigPath == "" {
		return fmt.Errorf("--keys is a required parameter for provisioning flow")
	}

	// Even if something crashed, things still were provisioned, meaning the cache should stay
	configCache := make(config.BoyarConfigCache)

	if flags.daemonize {
		supervized.GoForever(func() {
			for {
				if err := boyar.ReportStatus(context.Background(), logger); err != nil {
					logger.Error("status check failed", log.Error(err))
				}
				<-time.After(1 * time.Minute)
			}
		})

		<-supervized.GoForever(func() {
			for first := true; ; first = false {
				cfg, err := config.GetConfiguration(flags.configUrl, flags.ethereumEndpoint, flags.topologyContractAddress, flags.keyPairConfigPath)
				if err != nil {
					logger.Error("invalid configuration", log.Error(err))
				} else {
					// skip delay when provisioning for the first time when the node goes up
					if !first {
						reloadTimeDelay := cfg.ReloadTimeDelay(flags.maxReloadTimeDelay)
						logger.Info("waiting to apply new configuration", log.String("delay", flags.maxReloadTimeDelay.String()))
						<-time.After(reloadTimeDelay)
					}

					ctx, cancel := context.WithTimeout(context.Background(), flags.timeout)
					defer cancel()

					boyar.FullFlow(ctx, cfg, configCache, logger)
				}

				<-time.After(flags.pollingInterval)
			}
		})
	} else {
		cfg, err := config.GetConfiguration(flags.configUrl, flags.ethereumEndpoint, flags.topologyContractAddress, flags.keyPairConfigPath)
		if err != nil {
			return fmt.Errorf("invalid configuration: %s", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), flags.timeout)
		defer cancel()

		return boyar.FullFlow(ctx, cfg, configCache, logger)
	}

	return nil
}
