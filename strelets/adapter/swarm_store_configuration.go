package adapter

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"strings"
)

const NGINX_CONF = "nginx.conf"
const VCHAINS_CONF = "vchains.conf"
const SSL_CERT = "ssl-cert"
const SSL_KEY = "ssl-key"

func (d *dockerSwarm) storeVirtualChainConfiguration(ctx context.Context, containerName string, config *AppConfig) (*dockerSwarmSecretsConfig, error) {
	secrets := &dockerSwarmSecretsConfig{}

	if configSecretId, err := d.saveSwarmSecret(ctx, containerName, "config", config.Config); err != nil {
		return nil, fmt.Errorf("could not store config secret: %s", err)
	} else {
		secrets.configSecretId = configSecretId
	}

	if keyPairSecretId, err := d.saveSwarmSecret(ctx, containerName, "keyPair", config.KeyPair); err != nil {
		return nil, fmt.Errorf("could not store key pair secret: %s", err)
	} else {
		secrets.keysSecretId = keyPairSecretId
	}

	if networkSecretId, err := d.saveSwarmSecret(ctx, containerName, "network", config.Network); err != nil {
		return nil, fmt.Errorf("could not store network config secret: %s", err)
	} else {
		secrets.networkSecretId = networkSecretId
	}

	return secrets, nil
}

func (d *dockerSwarm) saveSwarmSecret(ctx context.Context, containerName string, secretName string, content []byte) (string, error) {
	secretId := getSwarmSecretName(containerName, secretName)

	if secrets, err := d.client.SecretList(ctx, types.SecretListOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{
			"name",
			secretId,
		}),
	}); err != nil {
		return "", fmt.Errorf("could not list swarm secrets: %s", err)
	} else {
		for _, secret := range secrets {
			if err := d.client.SecretRemove(ctx, secret.ID); err != nil {
				return "", fmt.Errorf("failed to remove a secret %s: :%s", secretName, err)
			}
		}
	}

	secretSpec := swarm.SecretSpec{
		Data: content,
	}
	secretSpec.Name = secretId

	response, err := d.client.SecretCreate(ctx, secretSpec)
	return response.ID, err
}

func getSwarmSecretName(containerName string, secretName string) string {
	return strings.Join([]string{containerName, secretName}, "-")
}

func (d *dockerSwarm) storeNginxConfiguration(ctx context.Context, config *ReverseProxyConfig) (*dockerSwarmNginxSecretsConfig, error) {
	secrets := &dockerSwarmNginxSecretsConfig{}

	if nginxConfId, err := d.saveSwarmSecret(ctx, PROXY_CONTAINER_NAME, NGINX_CONF, []byte(DEFAULT_NGINX_CONFIG)); err != nil {
		return nil, fmt.Errorf("could not store nginx default config secret: %s", err)
	} else {
		secrets.nginxConfId = nginxConfId
	}

	if vchainConfId, err := d.saveSwarmSecret(ctx, PROXY_CONTAINER_NAME, VCHAINS_CONF, []byte(config.NginxConfig)); err != nil {
		return nil, fmt.Errorf("could not store nginx vchains config secret: %s", err)
	} else {
		secrets.vchainConfId = vchainConfId
	}

	if config.SSLCertificate != nil {
		if sslCertificateId, err := d.saveSwarmSecret(ctx, PROXY_CONTAINER_NAME, SSL_CERT, config.SSLCertificate); err != nil {
			return nil, fmt.Errorf("could not store nginx ssl certificate secret: %s", err)
		} else {
			secrets.sslCertificateId = sslCertificateId
		}

	}

	if config.SSLPrivateKey != nil {
		if sslPrivateKeyId, err := d.saveSwarmSecret(ctx, PROXY_CONTAINER_NAME, SSL_KEY, config.SSLPrivateKey); err != nil {
			return nil, fmt.Errorf("could not store nginx ssl private key secret: %s", err)
		} else {
			secrets.sslPrivateKeyId = sslPrivateKeyId
		}
	}

	return secrets, nil
}

const DEFAULT_NGINX_CONFIG = `
daemon off;

user  nginx;
worker_processes  1;

error_log  /var/log/nginx/error.log warn;
pid        /var/run/nginx.pid;


events {
    worker_connections  1024;
}


http {
    include       /etc/nginx/mime.types;
    default_type  application/octet-stream;

    log_format  main  '$remote_addr - $remote_user [$time_local] "$request" '
                      '$status $body_bytes_sent "$http_referer" '
                      '"$http_user_agent" "$http_x_forwarded_for"';

    access_log  /var/log/nginx/access.log  main;

    sendfile        on;
    #tcp_nopush     on;

    keepalive_timeout  65;

    #gzip  on;

    include /var/run/secrets/vchains.conf;
}
`
