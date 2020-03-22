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

	namespace := "node123-proxy"
	httpPort := uint32(80)
	sslPort := uint32(443)

	secrets := &dockerSwarmNginxSecretsConfig{
		vchainConfId: "vchain-config-id",
		nginxConfId:  "nginx-config-id",
	}
	spec := getNginxServiceSpec(namespace, httpPort, sslPort, secrets, nil, nil)

	require.EqualValues(t, "node123-proxy-stack", spec.Name)

	require.EqualValues(t, swarm.TaskSpec{
		ContainerSpec: &swarm.ContainerSpec{
			Image: "nginx:latest",
			Command: []string{
				"nginx",
				"-c", "/var/run/secrets/nginx.conf",
			},
			Secrets: []*swarm.SecretReference{
				{
					SecretName: "node123-proxy-nginx.conf",
					SecretID:   "nginx-config-id",
					File: &swarm.SecretReferenceFileTarget{
						Name: "nginx.conf",
						UID:  "0",
						GID:  "0",
					},
				},
				{
					SecretName: "node123-proxy-vchains.conf",
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
	}, spec.TaskTemplate)

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
