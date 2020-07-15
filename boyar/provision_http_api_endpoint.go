package boyar

import (
	"context"
	"fmt"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/scribe/log"
	"io/ioutil"
)

func (b *boyar) ProvisionHttpAPIEndpoint(ctx context.Context) error {
	b.nginxLock.Lock()
	defer b.nginxLock.Unlock()

	if b.cache.nginx.CheckNewJsonValue(getNginxConfig(b.config)) {
		sslOptions := b.config.SSLOptions()
		sslEnabled := sslOptions.SSLCertificatePath != "" && sslOptions.SSLPrivateKeyPath != ""

		config := &adapter.ReverseProxyConfig{
			ContainerName: b.config.NamespacedContainerName(adapter.PROXY_CONTAINER_NAME),
			NodeAddress:   string(b.config.NodeAddress()),
			NginxConfig:   getNginxConfig(b.config),
			HTTPPort:      b.config.OrchestratorOptions().HTTPPort,
			SSLPort:       b.config.OrchestratorOptions().SSLPort,
			Services:      getReverseProxyServices(b.config),
		}

		if sslEnabled {
			if sslCertificate, err := ioutil.ReadFile(sslOptions.SSLCertificatePath); err != nil {
				return fmt.Errorf("could not read SSL certificate from %s: %s", sslOptions.SSLCertificatePath, err)
			} else {
				config.SSLCertificate = sslCertificate
			}

			if sslPrivateKey, err := ioutil.ReadFile(sslOptions.SSLPrivateKeyPath); err != nil {
				return fmt.Errorf("could not read SSL private key from %s: %s", sslOptions.SSLCertificatePath, err)
			} else {
				config.SSLPrivateKey = sslPrivateKey
			}
		}

		if err := b.orchestrator.RunReverseProxy(ctx, config); err != nil {
			b.logger.Error("failed to apply http proxy configuration", log.Error(err))
			b.cache.nginx.Clear()
			return err
		}

		b.logger.Info("updated http proxy configuration")
	}
	return nil
}

func getReverseProxyServices(cfg config.NodeConfiguration) (services []adapter.ReverseProxyConfigService) {
	for serviceName, _ := range cfg.Services() {
		services = append(services, adapter.ReverseProxyConfigService{
			Name:        serviceName,
			ServiceName: cfg.NamespacedContainerName(serviceName),
		})
	}

	// both services and vchains use the same volumes for logs and status
	for _, vchain := range cfg.Chains() {
		services = append(services, adapter.ReverseProxyConfigService{
			Name:        vchain.GetContainerName(),
			ServiceName: cfg.NamespacedContainerName(vchain.GetContainerName()),
		})
	}

	// special case to pass boyar logs from the outside	                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     \'
	services = append(services, adapter.ReverseProxyConfigService{
		Name:        BOYAR_SERVICE,
		ServiceName: BOYAR_SERVICE,
	})

	return
}
