package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/orbs-network/boyarin/boyar"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/services"
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

	_ = flag.Bool("daemonize", true, "DEPRECATED (always true)")
	pollingIntervalPtr := flag.Duration("polling-interval", 1*time.Minute, "how often to poll for configuration in daemon mode (duration: 1s, 1m, 1h, etc)")
	maxReloadTimePtr := flag.Duration("max-reload-time-delay", 15*time.Minute, "introduces jitter to reloading configuration to make network more stable, only works in daemon mode (duration: 1s, 1m, 1h, etc)")

	timeoutPtr := flag.Duration("timeout", 10*time.Minute, "timeout for provisioning all virtual chains (duration: 1s, 1m, 1h, etc)")

	ethereumEndpointPtr := flag.String("ethereum-endpoint", "", "Ethereum endpoint")
	topologyContractAddressPtr := flag.String("topology-contract-address", "", "Ethereum address for topology contract")

	loggerHttpEndpointPtr := flag.String("logger-http-endpoint", "", "Logz.io http endpoint")
	logFilePath := flag.String("log", "", "path to log file")

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
		LogFilePath:             *logFilePath,
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

	logger, err := config.GetLogger(flags)
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

	services.WatchAndReportServicesStatus(logger)

	runCoreBoyarService(flags, logger)

	return nil
}

const BOYAR_CONFIG_HASH = "boyar config"

func runCoreBoyarService(flags *config.Flags, logger log.Logger) {
	// Even if something crashed, things still were provisioned, meaning the cache should stay
	configCache := config.NewCache()
	<-supervized.GoForever(func(first bool) {
		defer func() {
			<-time.After(flags.PollingInterval)
		}()
		cfg, err := config.GetConfiguration(flags)
		if err != nil {
			logger.Error("invalid configuration", log.Error(err))
			return
		}
		hash := cfg.Hash()
		if hash == configCache.Get(BOYAR_CONFIG_HASH) {
			logger.Error("configuration has not changed")
			return
		}
		// random delay when provisioning change (that is, not bootstrap flow)
		if !first {
			randomDelay(cfg, flags.MaxReloadTimeDelay, logger)
		}
		ctx, cancel := context.WithTimeout(context.Background(), flags.Timeout)
		defer cancel()

		err = boyar.Flow(ctx, cfg, configCache, logger)
		if err != nil {
			logger.Error("error during execution", log.Error(err))
		}
		configCache.Put(BOYAR_CONFIG_HASH, hash)
	})
}

func randomDelay(cfg config.NodeConfiguration, maxDelay time.Duration, logger log.Logger) {
	reloadTimeDelay := cfg.ReloadTimeDelay(maxDelay)
	logger.Info("waiting to apply new configuration", log.String("delay", maxDelay.String()))
	<-time.After(reloadTimeDelay)
}
