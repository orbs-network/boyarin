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

	EthereumEndpoint string

	LoggerHttpEndpoint string
	LogFilePath        string

	StatusFilePath  string
	MetricsFilePath string

	OrchestratorOptions string

	ManagementConfig string

	// Autoupdate
	AutoUpdate          bool
	ShutdownAfterUpdate bool
	BoyarBinaryPath     string

	// Testing only
	WithNamespace bool
}
