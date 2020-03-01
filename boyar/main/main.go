package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/services"
	"github.com/orbs-network/boyarin/strelets/adapter"
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

	if *help {
		flag.Usage()
		return
	}

	logger, err := config.GetLogger(flags)
	if err != nil {
		os.Exit(1)
	}

	if *showConfiguration {
		printConfiguration(flags, logger)
		return
	}
	waiter, err := services.Execute(context.Background(), flags, logger)
	if err != nil {
		logger.Error("Startup failure", log.Error(err))
		os.Exit(1)
	}
	// should block forever
	waiter.WaitUntilShutdown(context.Background())
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
