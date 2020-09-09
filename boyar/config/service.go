package config

type Service struct {
	InternalPort int
	ExternalPort int
	DockerConfig DockerConfig
	Config       map[string]interface{}
	Disabled     bool

	InjectNodePrivateKey  bool
	ExecutablePath        string
	AllowAccessToSigner   bool
	AllowAccessToServices bool

	MountNodeLogs bool
}

type Services map[string]*Service

func (s Services) Signer() *Service {
	return s["signer"]
}

func (s Services) Management() *Service {
	return s["management-service"]
}

func (s Services) Names() (names []string) {
	for name, _ := range s {
		names = append(names, name)
	}

	return
}

const SIGNER = "signer"
