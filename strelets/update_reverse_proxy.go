package strelets

import (
	"context"
	"fmt"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"io/ioutil"
)

type UpdateReverseProxyInput struct {
	Chains []*VirtualChain
	IP     string

	SSLOptions adapter.SSLOptions
}

func (s *strelets) UpdateReverseProxy(ctx context.Context, input *UpdateReverseProxyInput) error {
	sslEnabled := input.SSLOptions.SSLCertificatePath != "" && input.SSLOptions.SSLPrivateKeyPath != ""

	config := &adapter.ReverseProxyConfig{
		NginxConfig: getNginxConfig(input.Chains, input.IP, sslEnabled),
	}

	if sslEnabled {
		if sslCertificate, err := ioutil.ReadFile(input.SSLOptions.SSLCertificatePath); err != nil {
			return fmt.Errorf("could not read SSL certificate from %s: %s", input.SSLOptions.SSLCertificatePath, err)
		} else {
			config.SSLCertificate = sslCertificate
		}

		if sslPrivateKey, err := ioutil.ReadFile(input.SSLOptions.SSLPrivateKeyPath); err != nil {
			return fmt.Errorf("could not read SSL private key from %s: %s", input.SSLOptions.SSLCertificatePath, err)
		} else {
			config.SSLPrivateKey = sslPrivateKey
		}
	}

	return s.orchestrator.RunReverseProxy(ctx, config)
}
