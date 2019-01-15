package adapter

import (
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/swarm"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func Test_getServiceSpec(t *testing.T) {
	containerName := "node1-vchain-42"
	secrets := []*swarm.SecretReference{
		getSecretReference(containerName, "some-secret-id", "some-secret-name", "some-secret.json"),
	}
	mounts := []mount.Mount{
		{Source: "vol1"},
		{Source: "vol2"},
	}

	restartDelay := time.Duration(10 * time.Second)
	replicas := uint64(1)

	spec := getVirtualChainServiceSpec("orbs:export", containerName, 16160, 8800, secrets, mounts)

	require.EqualValues(t, spec.Name, "stack-"+containerName)

	require.EqualValues(t, spec.TaskTemplate, swarm.TaskSpec{
		ContainerSpec: &swarm.ContainerSpec{
			Image: "orbs:export",
			Command: []string{
				"/opt/orbs/orbs-node",
				"--silent",
				"--log", "/opt/orbs/logs/node.log",
				"--config", "/var/run/secrets/some-secret.json",
			},
			Secrets: secrets,
			Sysctls: getSysctls(),
			Mounts:  mounts,
		},
		RestartPolicy: &swarm.RestartPolicy{
			Condition: "",
			Delay:     &restartDelay,
		},
	})

	require.EqualValues(t, spec.EndpointSpec, &swarm.EndpointSpec{
		Ports: []swarm.PortConfig{
			{
				Protocol:      "tcp",
				PublishMode:   swarm.PortConfigPublishModeIngress,
				PublishedPort: 16160,
				TargetPort:    8080,
			},
			{
				Protocol:      "tcp",
				PublishMode:   swarm.PortConfigPublishModeIngress,
				PublishedPort: 8800,
				TargetPort:    4400,
			},
		},
	})

	require.EqualValues(t, spec.Mode, swarm.ServiceMode{
		Replicated: &swarm.ReplicatedService{
			Replicas: &replicas,
		},
	})
}
