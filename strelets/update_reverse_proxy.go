package strelets

import (
	"context"
	"github.com/orbs-network/boyarin/strelets/adapter"
)

type UpdateReverseProxyInput struct {
	Chains []*VirtualChain
	IP     string

	SSLOptions adapter.SSLOptions
}

func (s *strelets) UpdateReverseProxy(ctx context.Context, input *UpdateReverseProxyInput) error {
	panic("removed")
}
