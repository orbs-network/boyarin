package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/orbs-network/boyarin/boyar"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/supervized"
	"github.com/orbs-network/boyarin/version"
	"github.com/orbs-network/scribe/log"
	"os"
	"time"
)

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

	orchestratorOptionsPtr := flag.String("orchestrator-options", "", "allows to override `orchestrator` section of boyar config, takes JSON object as a parameter")

	sslCertificatePathPtr := flag.String("ssl-certificate", "", "SSL certificate")
	sslPrivateKeyPtr := flag.String("ssl-private-key", "", "SSL private key")

	showConfiguration := flag.Bool("show-configuration", false, "show configuration and exit")
	help := flag.Bool("help", false, "show usage")
	showVersion := flag.Bool("version", false, "show version")

	flag.Parse()

	if *showVersion {
		fmt.Println(version.GetVersion().String())
		fmt.Println("Docker API version", adapter.DOCKER_API_VERSION)
		return
	}

	flags := &config.Flags{
		ConfigUrl:               *configUrlPtr,
		KeyPairConfigPath:       *keyPairConfigPathPtr,
		Daemonize:               *daemonizePtr,
		PollingInterval:         *pollingIntervalPtr,
		Timeout:                 *timeoutPtr,
		MaxReloadTimeDelay:      *maxReloadTimePtr,
		EthereumEndpoint:        *ethereumEndpointPtr,
		TopologyContractAddress: *topologyContractAddressPtr,
		LoggerHttpEndpoint:      *loggerHttpEndpointPtr,
		OrchestratorOptions:     *orchestratorOptionsPtr,
		SSLCertificatePath:      *sslCertificatePathPtr,
		SSLPrivateKeyPath:       *sslPrivateKeyPtr,
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

func getLogger(flags *config.Flags) (log.Logger, error) {
	outputs := []log.Output{log.NewFormattingOutput(os.Stdout, log.NewHumanReadableFormatter())}

	if flags.LoggerHttpEndpoint != "" {
		outputs = append(outputs, log.NewBulkOutput(
			log.NewHttpWriter(flags.LoggerHttpEndpoint),
			log.NewJsonFormatter().WithTimestampColumn("@timestamp"), 1))
	}

	logger := log.GetLogger().
		WithTags(log.String("app", "boyar")).
		WithOutput(outputs...).
		WithSourcePrefix("boyarin/")

	cfg, _ := config.NewStringConfigurationSource("{}", "")
	cfg.SetKeyConfigPath(flags.KeyPairConfigPath)
	if err := cfg.VerifyConfig(); err != nil {
		logger.Error("Invalid configuration", log.Error(err))
		return nil, err
	}

	return logger.WithTags(log.Node(string(cfg.NodeAddress()))), nil
}

func printConfiguration(flags *config.Flags, logger log.Logger) {
	cfg, err := config.GetConfiguration(flags)
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

func execute(flags *config.Flags, logger log.Logger) error {
	if flags.ConfigUrl == "" {
		return fmt.Errorf("--config-url is a required parameter for provisioning flow")
	}

	if flags.KeyPairConfigPath == "" {
		return fmt.Errorf("--keys is a required parameter for provisioning flow")
	}

	// Even if something crashed, things still were provisioned, meaning the cache should stay
	configCache := config.NewCache()

	if flags.Daemonize {
		supervized.GoForever(func() {
			for {
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				if err := boyar.ReportStatus(ctx, logger); err != nil {
					logger.Error("status check failed", log.Error(err))
				}
				cancel()
				<-time.After(1 * time.Minute)
			}
		})

		<-supervized.GoForever(func() {
			for first := true; ; first = false {
				cfg, err := config.GetConfiguration(flags)
				if err != nil {
					logger.Error("invalid configuration", log.Error(err))
				} else {
					// skip delay when provisioning for the first time when the node goes up
					if !first {
						reloadTimeDelay := cfg.ReloadTimeDelay(flags.MaxReloadTimeDelay)
						logger.Info("waiting to apply new configuration", log.String("delay", flags.MaxReloadTimeDelay.String()))
						<-time.After(reloadTimeDelay)
					}

					ctx, cancel := context.WithTimeout(context.Background(), flags.Timeout)
					defer cancel()

					boyar.Flow(ctx, cfg, configCache, logger)
				}

				<-time.After(flags.PollingInterval)
			}
		})
	} else {
		cfg, err := config.GetConfiguration(flags)
		if err != nil {
			return fmt.Errorf("invalid configuration: %s", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), flags.Timeout)
		defer cancel()

		return boyar.Flow(ctx, cfg, configCache, logger)
	}

	return nil
}
