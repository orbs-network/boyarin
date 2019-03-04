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

	serviceConfig := &ServiceConfig{
		ImageName:     "orbs:export",
		ContainerName: containerName,
		GossipPort:    8800,
		HttpPort:      16160,
	}

	spec := getVirtualChainServiceSpec(serviceConfig, secrets, mounts)

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
		Resources: &swarm.ResourceRequirements{
			Limits: &swarm.Resources{
				NanoCPUs:    1 * CPU_SHARES,
				MemoryBytes: 3000 * MEGABYTE,
			},
			Reservations: &swarm.Resources{
				NanoCPUs:    0.25 * CPU_SHARES,
				MemoryBytes: 300 * MEGABYTE,
			},
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

func Test_getResourceRequirements(t *testing.T) {
	defaultResourceRequirements := getResourceRequirements(0, 0, 0, 0)
	require.EqualValues(t, 3000*1024*1024, defaultResourceRequirements.Limits.MemoryBytes)
	require.EqualValues(t, 300*1024*1024, defaultResourceRequirements.Reservations.MemoryBytes)

	require.EqualValues(t, 1*10000000000, defaultResourceRequirements.Limits.NanoCPUs)
	require.EqualValues(t, int64(0.25*10000000000), defaultResourceRequirements.Reservations.NanoCPUs)

	limitMemory := getResourceRequirements(100, 0, 0, 0)
	require.EqualValues(t, 100*1024*1024, limitMemory.Limits.MemoryBytes)

	reserveMemory := getResourceRequirements(0, 0, 125, 0)
	require.EqualValues(t, 125*1024*1024, reserveMemory.Reservations.MemoryBytes)

	limitCPU := getResourceRequirements(0, 0.75, 0, 0)
	require.EqualValues(t, int64(0.75*10000000000), limitCPU.Limits.NanoCPUs)

	reserveCPU := getResourceRequirements(0, 0, 0, 2)
	require.EqualValues(t, 2*10000000000, reserveCPU.Reservations.NanoCPUs)
}
