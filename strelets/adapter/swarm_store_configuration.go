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
const SSL_CERT = "ssl-cert"
const SSL_KEY = "ssl-key"

func (d *dockerSwarmOrchestrator) saveSwarmSecret(ctx context.Context, containerName string, secretName string, content []byte) (string, error) {
	secretId := getSwarmSecretName(containerName, secretName)

	if secrets, err := d.client.SecretList(ctx, types.SecretListOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{
			Key:   "name",
			Value: secretId,
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

func (d *dockerSwarmOrchestrator) storeNginxConfiguration(ctx context.Context, config *ReverseProxyConfig) (*dockerSwarmNginxSecretsConfig, error) {
	secrets := &dockerSwarmNginxSecretsConfig{}

	if nginxConfId, err := d.saveSwarmSecret(ctx, config.ContainerName, NGINX_CONF, []byte(DEFAULT_NGINX_CONFIG)); err != nil {
		return nil, fmt.Errorf("could not store nginx default config secret: %s", err)
	} else {
		secrets.nginxConfId = nginxConfId
	}

	if config.SSLCertificate != nil {
		if sslCertificateId, err := d.saveSwarmSecret(ctx, config.ContainerName, SSL_CERT, config.SSLCertificate); err != nil {
			return nil, fmt.Errorf("could not store nginx ssl certificate secret: %s", err)
		} else {
			secrets.sslCertificateId = sslCertificateId
		}

	}

	if config.SSLPrivateKey != nil {
		if sslPrivateKeyId, err := d.saveSwarmSecret(ctx, config.ContainerName, SSL_KEY, config.SSLPrivateKey); err != nil {
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
    default_type  application/json;

    log_format  main  '$remote_addr - $remote_user [$time_local] "$request" '
                      '$status $body_bytes_sent "$http_referer" '
                      '"$http_user_agent" "$http_x_forwarded_for"';

    access_log  /var/log/nginx/access.log  main;

    sendfile        on;
    #tcp_nopush     on;

    keepalive_timeout  65;

    #gzip  on;

}
`
