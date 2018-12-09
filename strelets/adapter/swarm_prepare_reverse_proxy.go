package adapter

import (
	"context"
	"github.com/docker/docker/api/types/swarm"
	"time"
)

func (d *dockerSwarm) PrepareReverseProxy(ctx context.Context, config string) (Runner, error) {
	return &dockerSwarmRunner{
		client: d.client,
		spec: func() (swarm.ServiceSpec, error) {
			storedSecrets, err := d.storeNginxConfiguration(ctx, config)
			if err != nil {
				return swarm.ServiceSpec{}, err
			}
			return getNginxServiceSpec(storedSecrets), nil
		},
	}, nil
}

func getNginxServiceSpec(storedSecrets *dockerSwarmNginxSecretsConfig) swarm.ServiceSpec {
	restartDelay := time.Duration(10 * time.Second)
	replicas := uint64(1)

	secrets := []*swarm.SecretReference{
		getSecretReference(PROXY_CONTAINER_NAME, storedSecrets.nginxConfId, "nginx.conf", "nginx.conf"),
		getSecretReference(PROXY_CONTAINER_NAME, storedSecrets.vchainConfId, "vchains.conf", "vchains.conf"),
	}

	spec := swarm.ServiceSpec{
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{
				Image:   "nginx:latest",
				Secrets: secrets,
				Command: []string{
					"nginx", "-c", "/var/run/secrets/nginx.conf",
				},
			},
			RestartPolicy: &swarm.RestartPolicy{
				Delay: &restartDelay,
			},
		},
		Mode: getServiceMode(replicas),
		EndpointSpec: &swarm.EndpointSpec{
			Ports: []swarm.PortConfig{
				{
					Protocol:      "tcp",
					PublishMode:   swarm.PortConfigPublishModeIngress,
					PublishedPort: uint32(80),
					TargetPort:    80,
				},
			},
		},
	}
	spec.Name = getServiceId(PROXY_CONTAINER_NAME)

	return spec
}
