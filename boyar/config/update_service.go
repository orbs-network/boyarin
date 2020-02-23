package config

import (
	"fmt"
	"github.com/orbs-network/boyarin/strelets"
	"github.com/orbs-network/boyarin/strelets/adapter"
)

type Service struct {
	Port         int
	DockerConfig strelets.DockerConfig
	Config       map[string]interface{}
	Disabled     bool
}

func (s *Service) GetContainerName(serviceName string) string {
	return fmt.Sprintf("%s-%s", s.DockerConfig.ContainerNamePrefix, serviceName)
}

func (s *Service) SignerInternalEndpoint() string {
	return fmt.Sprintf("%s:%d", adapter.GetServiceId(s.GetContainerName(SIGNER)), s.Port)
}

type Services struct {
	Signer *Service `json:"signer"`
	Config *Service `json:"config"`
}

func (s Services) SignerOn() bool {
	return s.Signer != nil && s.Signer.Disabled == false
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
