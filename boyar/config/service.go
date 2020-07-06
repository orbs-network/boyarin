package config

type Service struct {
	InternalPort int
	ExternalPort int
	DockerConfig DockerConfig
	Config       map[string]interface{}
	Disabled     bool
}

type Services map[string]*Service

func (s Services) Signer() *Service {
	return s["signer"]
}

func (s Services) Management() *Service {
	return s["management-service"]
}

const SIGNER = "signer"
const MANAGEMENT = "management-service"

var SIGNER_SERVICE_CONFIG = ServiceConfig{
	Name:                   SIGNER,
	NeedsKeys:              true,
	Executable:             "/opt/orbs/orbs-signer",
	SignerNetworkEnabled:   true,
	ServicesNetworkEnabled: false,
}

var CONFIG_SERVICE_CONFIG = ServiceConfig{
	Name:                   MANAGEMENT,
	NeedsKeys:              false,
	Executable:             "/opt/orbs/service",
	SignerNetworkEnabled:   false,
	ServicesNetworkEnabled: true,
}

func NewServiceConfig(name string) ServiceConfig {
	return ServiceConfig{
		Name:                   name,
		NeedsKeys:              false,
		Executable:             "/opt/orbs/service",
		SignerNetworkEnabled:   false,
		ServicesNetworkEnabled: true,
	}
}

func (s Services) AsMap() map[ServiceConfig]*Service {
	result := map[ServiceConfig]*Service{
		SIGNER_SERVICE_CONFIG: s.Signer(),
		CONFIG_SERVICE_CONFIG: s.Management(),
	}

	for name, service := range s {
		if name != "signer" && name != "management-service" {
			result[NewServiceConfig(name)] = service
		}
	}

	return result
}

type ServiceConfig struct {
	Name                   string
	NeedsKeys              bool
	Executable             string
	SignerNetworkEnabled   bool
	ServicesNetworkEnabled bool
}
