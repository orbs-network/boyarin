package config

type Service struct {
	Port         int
	DockerConfig DockerConfig
	Config       map[string]interface{}
	Disabled     bool
}

type Services struct {
	Signer *Service `json:"signer"`
	Config *Service `json:"config"`
}

const SIGNER = "signer-service"
const CONFIG = "config-service"

var SIGNER_SERVICE_CONFIG = ServiceConfig{
	Name:       SIGNER,
	NeedsKeys:  true,
	External:   false,
	Executable: "/opt/orbs/orbs-signer",
}

var CONFIG_SERVICE_CONFIG = ServiceConfig{
	Name:       CONFIG,
	NeedsKeys:  false,
	External:   true,
	Executable: "/opt/orbs/service",
}

func (s Services) AsMap() map[ServiceConfig]*Service {
	return map[ServiceConfig]*Service{
		SIGNER_SERVICE_CONFIG: s.Signer,
		CONFIG_SERVICE_CONFIG: s.Config,
	}
}

type ServiceConfig struct {
	Name       string
	NeedsKeys  bool
	External   bool
	Executable string
}
