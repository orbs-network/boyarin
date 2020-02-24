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

func (s Services) AsMap() map[string]*Service {
	return map[string]*Service{
		SIGNER: s.Signer,
		CONFIG: s.Config,
	}
}

func (s Services) NeedsKeys(serviceId string) bool {
	switch serviceId {
	case SIGNER:
		return true
	}

	return false
}

type UpdateServiceInput struct {
	Name          string
	Service       *Service
	KeyPairConfig []byte `json:"-"` // Prevents possible key leak via log
}
