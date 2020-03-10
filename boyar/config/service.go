package config

type Service struct {
	Port         int
	ExternalPort int
	DockerConfig DockerConfig
	Config       map[string]interface{}
	Disabled     bool
}

type Services struct {
	Signer     *Service `json:"signer"`
	Management *Service `json:"management"`
}

const SIGNER = "signer-service"
const MANAGEMENT = "management-service"

var SIGNER_SERVICE_CONFIG = ServiceConfig{
	Name:       SIGNER,
	NeedsKeys:  true,
	Executable: "/opt/orbs/orbs-signer",
}

var CONFIG_SERVICE_CONFIG = ServiceConfig{
	Name:       MANAGEMENT,
	NeedsKeys:  false,
	Executable: "/opt/orbs/service",
}

func (s Services) AsMap() map[ServiceConfig]*Service {
	return map[ServiceConfig]*Service{
		SIGNER_SERVICE_CONFIG: s.Signer,
		CONFIG_SERVICE_CONFIG: s.Management,
	}
}

type ServiceConfig struct {
	Name       string
	NeedsKeys  bool
	Executable string
}
