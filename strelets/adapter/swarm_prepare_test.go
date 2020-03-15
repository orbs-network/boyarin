package adapter

import (
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/swarm"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func Test_getVirtualChainServiceSpec(t *testing.T) {
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
		ImageName:     "orbsnetwork/node:experimental",
		ContainerName: containerName,
		InternalPort:  8800,
		ExternalPort:  16160,
	}

	networkConfig := []swarm.NetworkAttachmentConfig{
		{Target: "signer"},
	}

	spec := getVirtualChainServiceSpec(serviceConfig, secrets, mounts, networkConfig)

	require.EqualValues(t, spec.Name, containerName+"-stack")

	require.EqualValues(t, spec.TaskTemplate, swarm.TaskSpec{
		ContainerSpec: &swarm.ContainerSpec{
			Image: "orbsnetwork/node:experimental",
			Command: []string{
				"/opt/orbs/orbs-node",
				"--silent",
				"--log", "/opt/orbs/logs/node.log",
				"--config", "/var/run/secrets/some-secret.json",
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
			Limits:       &swarm.Resources{},
			Reservations: &swarm.Resources{},
		},
	})

	require.EqualValues(t, spec.EndpointSpec, &swarm.EndpointSpec{
		Ports: []swarm.PortConfig{
			{
				Protocol:      "tcp",
				PublishMode:   swarm.PortConfigPublishModeIngress,
				PublishedPort: 16160,
				TargetPort:    8800,
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
	require.EqualValues(t, 0, defaultResourceRequirements.Limits.MemoryBytes)
	require.EqualValues(t, 0, defaultResourceRequirements.Reservations.MemoryBytes)

	require.EqualValues(t, 0, defaultResourceRequirements.Limits.NanoCPUs)
	require.EqualValues(t, 0, defaultResourceRequirements.Reservations.NanoCPUs)

	limitMemory := getResourceRequirements(100, 0, 0, 0)
	require.EqualValues(t, 100*1024*1024, limitMemory.Limits.MemoryBytes)

	reserveMemory := getResourceRequirements(0, 0, 125, 0)
	require.EqualValues(t, 125*1024*1024, reserveMemory.Reservations.MemoryBytes)

	limitCPU := getResourceRequirements(0, 0.75, 0, 0)
	require.EqualValues(t, int64(0.75*1000000000), limitCPU.Limits.NanoCPUs)

	reserveCPU := getResourceRequirements(0, 0, 0, 2)
	require.EqualValues(t, 2*1000000000, reserveCPU.Reservations.NanoCPUs)
}
