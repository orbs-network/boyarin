package strelets

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"io/ioutil"
)

type Service struct {
	Port         int
	DockerConfig DockerConfig
	Config       map[string]interface{}
	Disabled     bool
}

func (s *Service) getContainerName() string {
	return fmt.Sprintf("%s-signer-service", s.DockerConfig.ContainerNamePrefix)
}

func (s *Service) InternalEndpoint() string {
	return fmt.Sprintf("%s:%d", adapter.GetServiceId(s.getContainerName()), s.Port)
}

type Services struct {
	Signer *Service `json:"signer"`
}

func (s Services) SignerOn() bool {
	return s.Signer != nil && s.Signer.Disabled == false
}

type UpdateServiceInput struct {
	Service           *Service
	KeyPairConfigPath string
}

func (s *strelets) UpdateService(ctx context.Context, input *UpdateServiceInput) error {
	service := input.Service

	// FIXME pass actual []byte
	keyPair, err := ioutil.ReadFile(input.KeyPairConfigPath)
	if err != nil {
		return fmt.Errorf("could not read key pair config for signer service: %s at %s", err, input.KeyPairConfigPath)
	}

	serviceConfig := &adapter.ServiceConfig{
		ImageName:     service.DockerConfig.FullImageName(),
		ContainerName: service.getContainerName(),

		HttpPort: 7777,

		LimitedMemory:  service.DockerConfig.Resources.Limits.Memory,
		LimitedCPU:     service.DockerConfig.Resources.Limits.CPUs,
		ReservedMemory: service.DockerConfig.Resources.Reservations.Memory,
		ReservedCPU:    service.DockerConfig.Resources.Reservations.CPUs,
	}

	jsonConfig, _ := json.Marshal(service.Config)

	appConfig := &adapter.AppConfig{
		KeyPair: keyPair,
		Config:  jsonConfig,
	}

	if runner, err := s.orchestrator.PrepareService(ctx, serviceConfig, appConfig); err != nil {
		return err
	} else {
		return runner.Run(ctx)
	}
}
