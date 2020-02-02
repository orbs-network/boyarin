package adapter

import (
	"github.com/docker/docker/api/types/swarm"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func Test_getServiceSpec(t *testing.T) {
	containerName := "signer-service"
	secrets := []*swarm.SecretReference{
		getSecretReference(containerName, "some-secret-id", "some-secret-name", "some-secret.json"),
	}

	restartDelay := time.Duration(10 * time.Second)
	replicas := uint64(1)

	serviceConfig := &ServiceConfig{
		ImageName:     "orbs:signer",
		ContainerName: containerName,
	}

	networkConfig := []swarm.NetworkAttachmentConfig{
		{Target: "signer"},
	}

	spec := getServiceSpec(serviceConfig, secrets, networkConfig)

	require.EqualValues(t, spec.Name, containerName+"-stack")

	require.EqualValues(t, spec.TaskTemplate, swarm.TaskSpec{
		ContainerSpec: &swarm.ContainerSpec{
			Image: "orbs:signer",
			Command: []string{
				"/opt/orbs/orbs-signer",
				"--config", "/run/secrets/some-secret.json",
			},
			Secrets: secrets,
			Sysctls: GetSysctls(),
		},
		RestartPolicy: &swarm.RestartPolicy{
			Condition: "",
			Delay:     &restartDelay,
		},
		Resources: &swarm.ResourceRequirements{
			Limits:       &swarm.Resources{},
			Reservations: &swarm.Resources{},
		},
	})

	require.Nil(t, spec.EndpointSpec)

	require.EqualValues(t, spec.Mode, swarm.ServiceMode{
		Replicated: &swarm.ReplicatedService{
			Replicas: &replicas,
		},
	})
}
