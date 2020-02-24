package adapter

import (
	"context"
	"github.com/docker/docker/api/types/swarm"
	"time"
)

type ReverseProxyConfig struct {
	ContainerName  string
	NginxConfig    string
	SSLCertificate []byte
	SSLPrivateKey  []byte
}

func (d *dockerSwarmOrchestrator) RunReverseProxy(ctx context.Context, config *ReverseProxyConfig) error {
	serviceName := GetServiceId(config.ContainerName)
	if err := d.ServiceRemove(ctx, serviceName); err != nil {
		return err
	}

	storedSecrets, err := d.storeNginxConfiguration(ctx, config)
	if err != nil {
		return err
	}
	spec := getNginxServiceSpec(config.ContainerName, storedSecrets)
	return d.create(ctx, spec, "")
}

func getNginxServiceSpec(namespace string, storedSecrets *dockerSwarmNginxSecretsConfig) swarm.ServiceSpec {
	restartDelay := time.Duration(10 * time.Second)
	replicas := uint64(1)

	secrets := []*swarm.SecretReference{
		getSecretReference(namespace, storedSecrets.nginxConfId, NGINX_CONF, "nginx.conf"),
		getSecretReference(namespace, storedSecrets.vchainConfId, VCHAINS_CONF, "vchains.conf"),
	}

	if storedSecrets.sslCertificateId != "" {
		secrets = append(secrets, getSecretReference(namespace, storedSecrets.sslCertificateId, SSL_CERT, "ssl-cert"))
	}

	if storedSecrets.sslPrivateKeyId != "" {
		secrets = append(secrets, getSecretReference(namespace, storedSecrets.sslPrivateKeyId, SSL_KEY, "ssl-key"))
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
				Sysctls: GetSysctls(),
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
	spec.Name = GetServiceId(namespace)

	return spec
}
