package adapter

import (
	"github.com/docker/docker/api/types/swarm"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func Test_getNginxServiceSpec(t *testing.T) {
	restartDelay := time.Duration(10 * time.Second)
	replicas := uint64(1)

	secrets := &dockerSwarmNginxSecretsConfig{
		vchainConfId: "vchain-config-id",
		nginxConfId:  "nginx-config-id",
	}
	spec := getNginxServiceSpec(secrets)

	require.EqualValues(t, spec.Name, PROXY_CONTAINER_NAME+"-stack")

	require.EqualValues(t, spec.TaskTemplate, swarm.TaskSpec{
		ContainerSpec: &swarm.ContainerSpec{
			Image: "nginx:latest",
			Command: []string{
				"nginx",
				"-c", "/var/run/secrets/nginx.conf",
			},
			Secrets: []*swarm.SecretReference{
				{
					SecretName: getSwarmSecretName(PROXY_CONTAINER_NAME, "nginx.conf"),
					SecretID:   "nginx-config-id",
					File: &swarm.SecretReferenceFileTarget{
						Name: "nginx.conf",
						UID:  "0",
						GID:  "0",
					},
				},
				{
					SecretName: getSwarmSecretName(PROXY_CONTAINER_NAME, "vchains.conf"),
					SecretID:   "vchain-config-id",
					File: &swarm.SecretReferenceFileTarget{
						Name: "vchains.conf",
						UID:  "0",
						GID:  "0",
					},
				},
			},
			Sysctls: GetSysctls(),
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
				PublishedPort: 80,
				TargetPort:    80,
			},
		},
	})

	require.EqualValues(t, spec.Mode, swarm.ServiceMode{
		Replicated: &swarm.ReplicatedService{
			Replicas: &replicas,
		},
	})
}
