package adapter

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/swarm"
	"strings"
)

func (d *dockerSwarm) storeVirtualChainConfiguration(ctx context.Context, containerName string, config *AppConfig) (*dockerSwarmSecretsConfig, error) {
	secrets := &dockerSwarmSecretsConfig{}

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
	d.client.SecretRemove(ctx, secretId)

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

func (d *dockerSwarm) storeNginxConfiguration(ctx context.Context, config string) (*dockerSwarmNginxSecretsConfig, error) {
	secrets := &dockerSwarmNginxSecretsConfig{}

	if nginxConfId, err := d.saveSwarmSecret(ctx, PROXY_CONTAINER_NAME, "nginx.conf", []byte(DEFAULT_NGINX_CONFIG)); err != nil {
		return nil, fmt.Errorf("could not store nginx default config secret: %s", err)
	} else {
		secrets.nginxConfId = nginxConfId
	}

	if vchainConfId, err := d.saveSwarmSecret(ctx, PROXY_CONTAINER_NAME, "vchains.conf", []byte(config)); err != nil {
		return nil, fmt.Errorf("could not store nginx vchains config secret: %s", err)
	} else {
		secrets.vchainConfId = vchainConfId
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
