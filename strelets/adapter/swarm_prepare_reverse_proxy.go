package adapter

import (
	"context"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/swarm"
	"path"
	"time"
)

type ReverseProxyConfigService struct {
	Name        string
	ServiceName string
}

type ReverseProxyConfig struct {
	ContainerName string

	HTTPPort uint32
	SSLPort  uint32

	NginxConfig string

	Services []ReverseProxyConfigService

	SSLCertificate []byte
	SSLPrivateKey  []byte
}

const DEFAULT_HTTP_PORT = uint32(80)
const DEFAULT_SSL_PORT = uint32(443)

func (d *dockerSwarmOrchestrator) RunReverseProxy(ctx context.Context, config *ReverseProxyConfig) error {
	if err := d.RemoveService(ctx, config.ContainerName); err != nil {
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

	var networks []swarm.NetworkAttachmentConfig
	proxyNetwork, err := d.getNetwork(ctx, SHARED_PROXY_NETWORK)
	if err != nil {
		return err
	}
	networks = append(networks, proxyNetwork)

	var mounts []mount.Mount
	for _, nodeService := range config.Services {
		if ms, err := d.provisionStatusVolume(ctx, nodeService.ServiceName, getNginxStatusMountPath(nodeService.Name)); err != nil {
			return err
		} else {
			mounts = append(mounts, ms...)
		}
	}

	spec := getNginxServiceSpec(config.ContainerName, httpPort, sslPort, storedSecrets, networks, mounts)
	return d.create(ctx, spec, "")
}

func getNginxServiceSpec(namespace string, httpPort uint32, sslPort uint32, storedSecrets *dockerSwarmNginxSecretsConfig, networks []swarm.NetworkAttachmentConfig, mounts []mount.Mount) swarm.ServiceSpec {
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
				Mounts:  mounts,
			},
			RestartPolicy: &swarm.RestartPolicy{
				Delay: &restartDelay,
			},
		},
		Mode: getServiceMode(replicas),
		EndpointSpec: &swarm.EndpointSpec{
			Ports: ports,
		},
		Networks: networks,
	}
	spec.Name = namespace

	return spec
}

func getNginxStatusMountPath(simpleName string) string {
	return path.Join(ORBS_STATUS_TARGET, simpleName)
}
