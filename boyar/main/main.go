package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/recovery"
	"github.com/orbs-network/boyarin/services"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/version"
	"github.com/orbs-network/scribe/log"
)

func main() {
	basicLogger := log.GetLogger()
	basicLogger.Info("Boyar main version: " + version.GetVersion().Semantic)

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
	metricsFilePath := flag.String("metrics", "", "path to metrics file")

	orchestratorOptionsPtr := flag.String("orchestrator-options", "", "allows to override `orchestrator` section of boyar config, takes JSON object as a parameter")

	sslCertificatePathPtr := flag.String("ssl-certificate", "", "SSL certificate")
	sslPrivateKeyPtr := flag.String("ssl-private-key", "", "SSL private key")

	managementConfig := flag.String("management-config", "", "bootstrap only a configuration provider service and then retrieve all configuration from it")

	showConfiguration := flag.Bool("show-configuration", false, "print configuration and exit")
	help := flag.Bool("help", false, "show usage")
	showVersion := flag.Bool("version", false, "show version")

	showStatus := flag.Bool("show-status", false, "print status in json format and exit")

	autoUpdate := flag.Bool("auto-update", false, "enables boyar binary auto update")
	shutdownAfterUpdate := flag.Bool("shutdown-after-update", false, "the process shuts down after automatic update is performed and **DOES NOT** restart; recommended to be used with an external process manager")

	bootstrapResetTimeout := flag.Duration("bootstrap-reset-timeout", 0, "if the process is unable to receive valid configuration within a limited timeframe (duration: 1s, 1m, 1h, etc), it will exit with an error; recommended to be used with an external process manager, (default 0s, off)")

	flag.Parse()

	if *showVersion {
		fmt.Println(version.GetVersion().String())
		fmt.Println("Docker API version", adapter.DOCKER_API_VERSION)
		return
	}

	executable, _ := os.Executable()
	executableWithoutSymlink, _ := filepath.EvalSymlinks(executable)

	flags := &config.Flags{
		ConfigUrl:             *configUrlPtr,
		KeyPairConfigPath:     *keyPairConfigPathPtr,
		LogFilePath:           *logFilePath,
		StatusFilePath:        *statusFilePath,
		MetricsFilePath:       *metricsFilePath,
		PollingInterval:       *pollingIntervalPtr,
		Timeout:               *timeoutPtr,
		MaxReloadTimeDelay:    *maxReloadTimePtr,
		EthereumEndpoint:      *ethereumEndpointPtr,
		LoggerHttpEndpoint:    *loggerHttpEndpointPtr,
		OrchestratorOptions:   *orchestratorOptionsPtr,
		SSLCertificatePath:    *sslCertificatePathPtr,
		SSLPrivateKeyPath:     *sslPrivateKeyPtr,
		ManagementConfig:      *managementConfig,
		AutoUpdate:            *autoUpdate,
		ShutdownAfterUpdate:   *shutdownAfterUpdate,
		BoyarBinaryPath:       executableWithoutSymlink,
		BootstrapResetTimeout: *bootstrapResetTimeout,
	}

	if *showStatus {
		ctx, cancel := context.WithTimeout(context.Background(), services.SERVICE_STATUS_REPORT_TIMEOUT)
		defer cancel()

		status, _ := services.GetStatusAndMetrics(ctx, basicLogger, flags, time.Now(), services.SERVICE_STATUS_REPORT_TIMEOUT)
		rawJSON, _ := json.MarshalIndent(status, "  ", "  ")
		fmt.Println(string(rawJSON))

		return
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

	// start recovery //////////////////////////////
	logger.Info("============================================")
	keys, err := config.NewKeysConfig(flags.KeyPairConfigPath)
	nodeAddress := ""
	if err == nil {
		nodeAddress = keys.Address()
	} else {
		logger.Error("failed to get recovery node address, " + err.Error())
	}
	// generate random address if failed
	if nodeAddress == "" {
		nodeAddress = fmt.Sprintf("RAND_%d", rand.Intn(0xFFFFFFFF))
	}
	// go on to init recovery anyways
	logger.Info("recovery node address is: " + nodeAddress)
	url := fmt.Sprintf("https://deployment.orbs.network/boyar_recovery/node/0x%s/main.json", nodeAddress)
	// for testing
	//url := fmt.Sprintf("https://raw.githubusercontent.com/amihaz/staging-deployment/main/boyar_recovery/node/0x%s/main.json", keys.Address())
	recovConfig := recovery.Config{
		IntervalMinute: 60 * 6,
		TimeoutMinute:  30,
		Url:            url,
	}
	recovery.Init(recovConfig, logger)

	// start
	recovery.GetInstance().Start(true)
	logger.Info("============================================")

	// start services
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
