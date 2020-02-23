package boyar

import (
	"context"
	"fmt"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/orbs-network/scribe/log"
	"io/ioutil"
)

func (b *boyar) ProvisionHttpAPIEndpoint(ctx context.Context) error {
	b.nginxLock.Lock()
	defer b.nginxLock.Unlock()
	// TODO is there a better way to get a loopback interface?
	nginxConfig := getNginxCompositeConfig(b.config)

	if b.cache.nginx.CheckNewJsonValue(nginxConfig) {
		sslEnabled := nginxConfig.SSLOptions.SSLCertificatePath != "" && nginxConfig.SSLOptions.SSLPrivateKeyPath != ""

		config := &adapter.ReverseProxyConfig{
			NginxConfig: getNginxConfig(nginxConfig.Chains, nginxConfig.IP, sslEnabled),
		}

		if sslEnabled {
			if sslCertificate, err := ioutil.ReadFile(nginxConfig.SSLOptions.SSLCertificatePath); err != nil {
				return fmt.Errorf("could not read SSL certificate from %s: %s", nginxConfig.SSLOptions.SSLCertificatePath, err)
			} else {
				config.SSLCertificate = sslCertificate
			}

			if sslPrivateKey, err := ioutil.ReadFile(nginxConfig.SSLOptions.SSLPrivateKeyPath); err != nil {
				return fmt.Errorf("could not read SSL private key from %s: %s", nginxConfig.SSLOptions.SSLCertificatePath, err)
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

type UpdateReverseProxyInput struct {
	Chains []*config.VirtualChain
	IP     string

	SSLOptions adapter.SSLOptions
}

func getNginxCompositeConfig(cfg config.NodeConfiguration) *UpdateReverseProxyInput {
	return &UpdateReverseProxyInput{
		Chains:     cfg.Chains(),
		IP:         helpers.LocalIP(),
		SSLOptions: cfg.SSLOptions(),
	}
}
