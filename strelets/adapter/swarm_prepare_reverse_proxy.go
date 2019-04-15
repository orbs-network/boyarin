package adapter

import (
	"context"
	"github.com/docker/docker/api/types/swarm"
	"time"
)

type ReverseProxyConfig struct {
	NginxConfig    string
	SSLCertificate []byte
	SSLPrivateKey  []byte
}

func (d *dockerSwarm) PrepareReverseProxy(ctx context.Context, config *ReverseProxyConfig) (Runner, error) {
	return &dockerSwarmRunner{
		client: d.client,
		spec: func() (swarm.ServiceSpec, error) {
			storedSecrets, err := d.storeNginxConfiguration(ctx, config)
			if err != nil {
				return swarm.ServiceSpec{}, err
			}
			return getNginxServiceSpec(storedSecrets), nil
		},
		serviceName: getServiceId(PROXY_CONTAINER_NAME),
	}, nil
}

func getNginxServiceSpec(storedSecrets *dockerSwarmNginxSecretsConfig) swarm.ServiceSpec {
	restartDelay := time.Duration(10 * time.Second)
	replicas := uint64(1)

	secrets := []*swarm.SecretReference{
		getSecretReference(PROXY_CONTAINER_NAME, storedSecrets.nginxConfId, NGINX_CONF, "nginx.conf"),
		getSecretReference(PROXY_CONTAINER_NAME, storedSecrets.vchainConfId, VCHAINS_CONF, "vchains.conf"),
	}

	if storedSecrets.sslCertificateId != "" {
		secrets = append(secrets, getSecretReference(PROXY_CONTAINER_NAME, storedSecrets.sslCertificateId, SSL_CERT, "ssl-cert"))
	}

	if storedSecrets.sslPrivateKeyId != "" {
		secrets = append(secrets, getSecretReference(PROXY_CONTAINER_NAME, storedSecrets.sslPrivateKeyId, SSL_KEY, "ssl-key"))
	}

	ports := []swarm.PortConfig{
		{
			Protocol:      "tcp",
			PublishMode:   swarm.PortConfigPublishModeIngress,
			PublishedPort: uint32(80),
			TargetPort:    80,
		},
	}

	if storedSecrets.sslCertificateId != "" && storedSecrets.sslPrivateKeyId != "" {
		ports = append(ports, swarm.PortConfig{
			Protocol:      "tcp",
			PublishMode:   swarm.PortConfigPublishModeIngress,
			PublishedPort: uint32(443),
			TargetPort:    443,
		})
	}

	spec := swarm.ServiceSpec{
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{
				Image:   "nginx:latest",
				Secrets: secrets,
				Command: []string{
					"nginx", "-c", "/var/run/secrets/nginx.conf",
				},
				Sysctls: getSysctls(),
			},
			RestartPolicy: &swarm.RestartPolicy{
				Delay: &restartDelay,
			},
		},
		Mode: getServiceMode(replicas),
		EndpointSpec: &swarm.EndpointSpec{
			Ports: ports,
		},
	}
	spec.Name = getServiceId(PROXY_CONTAINER_NAME)

	return spec
}
