package config

type Service struct {
	InternalPort int
	ExternalPort int
	DockerConfig DockerConfig
	Config       map[string]interface{}
	Disabled     bool
}

type Services struct {
	Signer     *Service `json:"signer"`
	Management *Service `json:"management-service"`
	Ethereum   *Service `json:"ethereum"`
}

const SIGNER = "signer"
const MANAGEMENT = "management-service"
const ETHEREUM = "ethereum"

var SIGNER_SERVICE_CONFIG = ServiceConfig{
	Name:                 SIGNER,
	NeedsKeys:            true,
	Executable:           "/opt/orbs/orbs-signer",
	SignerNetworkEnabled: true,
}

var CONFIG_SERVICE_CONFIG = ServiceConfig{
	Name:                 MANAGEMENT,
	NeedsKeys:            false,
	Executable:           "/opt/orbs/service",
	SignerNetworkEnabled: false,
}

var ETHEREUM_SERVICE_CONFIG = ServiceConfig{
	Name:                 ETHEREUM,
	NeedsKeys:            false,
	Executable:           "/opt/orbs/service",
	SignerNetworkEnabled: false,
}

func (s Services) AsMap() map[ServiceConfig]*Service {
	return map[ServiceConfig]*Service{
		SIGNER_SERVICE_CONFIG:   s.Signer,
		CONFIG_SERVICE_CONFIG:   s.Management,
		ETHEREUM_SERVICE_CONFIG: s.Ethereum,
	}
}

type ServiceConfig struct {
	Name                 string
	NeedsKeys            bool
	Executable           string
	SignerNetworkEnabled bool
}
