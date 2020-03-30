package config

import "time"

type Flags struct {
	ConfigUrl         string
	KeyPairConfigPath string

	SSLCertificatePath string
	SSLPrivateKeyPath  string

	PollingInterval    time.Duration
	Timeout            time.Duration
	MaxReloadTimeDelay time.Duration

	EthereumEndpoint        string
	TopologyContractAddress string

	LoggerHttpEndpoint string
	LogFilePath        string

	OrchestratorOptions string

	ManagementConfig string

	// Testing only
	WithNamespace bool
}
