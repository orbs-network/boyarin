package strelets

import (
	"context"
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

type Services struct {
	Signer *Service `json:"signer"`
}

type UpdateServiceInput struct {
	Service           *Service
	KeyPairConfigPath string
}

func (s *strelets) UpdateService(ctx context.Context, input *UpdateServiceInput) error {
	service := input.Service
	keyPair, err := ioutil.ReadFile(input.KeyPairConfigPath)
	if err != nil {
		return fmt.Errorf("could not read key pair config for signer service: %s at %s", err, input.KeyPairConfigPath)
	}

	serviceConfig := &adapter.ServiceConfig{
		ImageName:     service.DockerConfig.FullImageName(),
		ContainerName: "signer-service",

		HttpPort: 7777,

		LimitedMemory:  service.DockerConfig.Resources.Limits.Memory,
		LimitedCPU:     service.DockerConfig.Resources.Limits.CPUs,
		ReservedMemory: service.DockerConfig.Resources.Reservations.Memory,
		ReservedCPU:    service.DockerConfig.Resources.Reservations.CPUs,
	}

	appConfig := &adapter.AppConfig{
		KeyPair: keyPair,
		// FIXME add proper config serialization
		Config: []byte("{}"),
	}

	if runner, err := s.orchestrator.PrepareService(ctx, serviceConfig, appConfig); err != nil {
		return err
	} else {
		return runner.Run(ctx)
	}
}

func (s Services) SignerOn() bool {
	return s.Signer != nil && s.Signer.Disabled == false
}
