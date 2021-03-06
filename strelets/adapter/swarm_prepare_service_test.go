package adapter

import (
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/swarm"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func Test_getServiceSpec(t *testing.T) {
	containerName := "signer"
	secrets := []*swarm.SecretReference{
		getSecretReference(containerName, "some-secret-id", "some-secret-name", "some-secret.json"),
	}

	restartDelay := time.Duration(10 * time.Second)
	replicas := uint64(1)

	serviceConfig := &ServiceConfig{
		ImageName:      "orbs:signer",
		ContainerName:  containerName,
		ExecutablePath: "/opt/orbs/orbs-signer",
	}

	networkConfig := []swarm.NetworkAttachmentConfig{
		{Target: "signer"},
	}

	mounts := []mount.Mount{
		{Source: "/tmp/a", Target: "/tmp/b"},
	}

	spec := getServiceSpec(serviceConfig, secrets, networkConfig, mounts)

	require.EqualValues(t, spec.Name, containerName)

	require.EqualValues(t, swarm.TaskSpec{
		ContainerSpec: &swarm.ContainerSpec{
			Image: "orbs:signer",
			Command: []string{
				"/opt/orbs/orbs-signer", "--config", "/run/secrets/some-secret.json",
			},
			Secrets: secrets,
			Sysctls: GetSysctls(),
			Mounts:  mounts,
		},
		RestartPolicy: &swarm.RestartPolicy{
			Condition: "",
			Delay:     &restartDelay,
		},
		Resources: &swarm.ResourceRequirements{
			Limits: &swarm.Resources{
				MemoryBytes: 3145728000,
				NanoCPUs:    1000000000,
			},
			Reservations: &swarm.Resources{},
		},
	}, spec.TaskTemplate)

	require.Nil(t, spec.EndpointSpec)

	require.EqualValues(t, spec.Mode, swarm.ServiceMode{
		Replicated: &swarm.ReplicatedService{
			Replicas: &replicas,
		},
	})
}
