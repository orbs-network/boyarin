package adapter

import (
	"context"
	"github.com/docker/docker/api/types/swarm"
	"time"
)

type ReverseProxyConfig struct {
	ContainerName string

	HTTPPort uint32
	SSLPort  uint32

	NginxConfig    string
	SSLCertificate []byte
	SSLPrivateKey  []byte
}

const DEFAULT_HTTP_PORT = uint32(80)
const DEFAULT_SSL_PORT = uint32(433)

func (d *dockerSwarmOrchestrator) RunReverseProxy(ctx context.Context, config *ReverseProxyConfig) error {
	serviceName := GetServiceId(config.ContainerName)
	if err := d.RemoveService(ctx, serviceName); err != nil {
		return err
	}

	storedSecrets, err := d.storeNginxConfiguration(ctx, config)
	if err != nil {
		return err
	}

	httpPort := DEFAULT_HTTP_PORT
	if config.HTTPPort != 0 {
		httpPort = config.HTTPPort
	}

	sslPort := DEFAULT_SSL_PORT
	if config.SSLPort != 0 {
		sslPort = config.SSLPort
	}

	spec := getNginxServiceSpec(config.ContainerName, httpPort, sslPort, storedSecrets)
	return d.create(ctx, spec, "")
}

func getNginxServiceSpec(namespace string, httpPort uint32, sslPort uint32, storedSecrets *dockerSwarmNginxSecretsConfig) swarm.ServiceSpec {
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
			PublishedPort: httpPort,
			TargetPort:    DEFAULT_HTTP_PORT,
		},
	}

	if storedSecrets.sslCertificateId != "" && storedSecrets.sslPrivateKeyId != "" {
		ports = append(ports, swarm.PortConfig{
			Protocol:      "tcp",
			PublishMode:   swarm.PortConfigPublishModeIngress,
			PublishedPort: sslPort,
			TargetPort:    DEFAULT_SSL_PORT,
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
