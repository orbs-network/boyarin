package config

import "time"

type Flags struct {
	ConfigUrl         string
	KeyPairConfigPath string

	Daemonize bool

	PollingInterval    time.Duration
	Timeout            time.Duration
	MaxReloadTimeDelay time.Duration

	EthereumEndpoint        string
	TopologyContractAddress string

	LoggerHttpEndpoint string

	OrchestratorOptions string
}
