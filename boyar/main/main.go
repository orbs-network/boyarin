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
	"github.com/prometheus/client_golang/prometheus"
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
	flag.String("topology-contract-address", "", "legacy parameter, will be removed in future versions")

	loggerHttpEndpointPtr := flag.String("logger-http-endpoint", "", "Logz.io http endpoint")
	logFilePath := flag.String("log", "", "path to log file")

	statusFilePath := flag.String("status", "", "path to status file")

	orchestratorOptionsPtr := flag.String("orchestrator-options", "", "allows to override `orchestrator` section of boyar config, takes JSON object as a parameter")

	sslCertificatePathPtr := flag.String("ssl-certificate", "", "SSL certificate")
	sslPrivateKeyPtr := flag.String("ssl-private-key", "", "SSL private key")

	managementConfig := flag.String("management-config", "", "bootstrap only a configuration provider service and then retrieve all configuration from it")

	showConfiguration := flag.Bool("show-configuration", false, "show configuration and exit")
	help := flag.Bool("help", false, "show usage")
	showVersion := flag.Bool("version", false, "show version")

	metricsOnly := flag.Bool("metrics-only", false, "print the list of prometheus metrics")
	statusOnly := flag.Bool("status-only", false, "print status in json format")

	flag.Parse()

	if *showVersion {
		fmt.Println(version.GetVersion().String())
		fmt.Println("Docker API version", adapter.DOCKER_API_VERSION)
		return
	}

	basicLogger := log.GetLogger()

	if *statusOnly {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if status, err := services.GetStatus(ctx, basicLogger, 30*time.Second); err != nil {
			basicLogger.Error("status check failed", log.Error(err))
			os.Exit(1)
		} else {
			rawJSON, _ := json.MarshalIndent(status, "  ", "  ")
			fmt.Println(string(rawJSON))
		}

		return
	}

	if *metricsOnly {
		registry := prometheus.NewRegistry()
		metrics, err := services.InitializeMetrics(registry)

		if err != nil {
			basicLogger.Error("failed to initialize metrics", log.Error(err))
			os.Exit(1)
		}

		services.CollectMetrics(metrics, basicLogger)

		serializedMetrics, err := services.GetSerializedMetrics(registry)
		if err != nil {
			basicLogger.Error("failed to serialize metrics", log.Error(err))
			os.Exit(1)
		}

		fmt.Println(serializedMetrics)

		return
	}

	flags := &config.Flags{
		ConfigUrl:           *configUrlPtr,
		KeyPairConfigPath:   *keyPairConfigPathPtr,
		LogFilePath:         *logFilePath,
		StatusFilePath:      *statusFilePath,
		PollingInterval:     *pollingIntervalPtr,
		Timeout:             *timeoutPtr,
		MaxReloadTimeDelay:  *maxReloadTimePtr,
		EthereumEndpoint:    *ethereumEndpointPtr,
		LoggerHttpEndpoint:  *loggerHttpEndpointPtr,
		OrchestratorOptions: *orchestratorOptionsPtr,
		SSLCertificatePath:  *sslCertificatePathPtr,
		SSLPrivateKeyPath:   *sslPrivateKeyPtr,
		ManagementConfig:    *managementConfig,
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

	if flags.ManagementConfig != "" {
		flags, err = services.Bootstrap(context.Background(), flags, logger)
		if err != nil {
			logger.Error("Bootstrapping failure", log.Error(err))
			os.Exit(1)
		}
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
