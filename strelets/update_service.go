package strelets

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/orbs-network/boyarin/strelets/adapter"
)

type Service struct {
	Port         int
	DockerConfig DockerConfig
	Config       map[string]interface{}
	Disabled     bool
}

func (s *Service) getContainerName(serviceName string) string {
	return fmt.Sprintf("%s-%s", s.DockerConfig.ContainerNamePrefix, serviceName)
}

func (s *Service) SignerInternalEndpoint() string {
	return fmt.Sprintf("%s:%d", adapter.GetServiceId(s.getContainerName(SIGNER)), s.Port)
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

func (s *strelets) UpdateService(ctx context.Context, input *UpdateServiceInput) error {
	service := input.Service
	imageName := service.DockerConfig.FullImageName()

	if service.Disabled {
		return fmt.Errorf("signer service is disabled")
	}

	if service.DockerConfig.Pull {
		if err := s.orchestrator.PullImage(ctx, imageName); err != nil {
			return fmt.Errorf("could not pull docker image: %s", err)
		}
	}

	serviceConfig := &adapter.ServiceConfig{
		ImageName:     imageName,
		ContainerName: service.getContainerName(input.Name),

		LimitedMemory:  service.DockerConfig.Resources.Limits.Memory,
		LimitedCPU:     service.DockerConfig.Resources.Limits.CPUs,
		ReservedMemory: service.DockerConfig.Resources.Reservations.Memory,
		ReservedCPU:    service.DockerConfig.Resources.Reservations.CPUs,
	}

	jsonConfig, _ := json.Marshal(service.Config)

	appConfig := &adapter.AppConfig{
		KeyPair: input.KeyPairConfig,
		Config:  jsonConfig,
	}

	return s.orchestrator.RunService(ctx, serviceConfig, appConfig)
}
