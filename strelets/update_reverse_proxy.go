package strelets

import (
	"context"
	"fmt"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"io/ioutil"
)

type UpdateReverseProxyInput struct {
	Chains             []*VirtualChain
	IP                 string
	SSLCertificatePath string
	SSLPrivateKeyPath  string
}

func (s *strelets) UpdateReverseProxy(ctx context.Context, input *UpdateReverseProxyInput) error {
	sslEnabled := input.SSLCertificatePath != "" && input.SSLPrivateKeyPath != ""

	config := &adapter.ReverseProxyConfig{
		NginxConfig: getNginxConfig(input.Chains, input.IP, sslEnabled),
	}

	if sslEnabled {
		if sslCertificate, err := ioutil.ReadFile(input.SSLCertificatePath); err != nil {
			return fmt.Errorf("could not read SSL certificate from %s: %s", input.SSLCertificatePath, err)
		} else {
			config.SSLCertificate = sslCertificate
		}

		if sslPrivateKey, err := ioutil.ReadFile(input.SSLPrivateKeyPath); err != nil {
			return fmt.Errorf("could not read SSL private key from %s: %s", input.SSLCertificatePath, err)
		} else {
			config.SSLPrivateKey = sslPrivateKey
		}
	}

	if runner, err := s.orchestrator.PrepareReverseProxy(ctx, config); err != nil {
		return err
	} else {
		return runner.Run(ctx)
	}
}
