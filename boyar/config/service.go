package config

type Service struct {
	InternalPort int
	ExternalPort int
	DockerConfig DockerConfig
	Config       map[string]interface{}
	Disabled     bool

	NeedsKeys              bool
	Executable             string
	SignerNetworkEnabled   bool
	ServicesNetworkEnabled bool
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
