package strelets

import "context"

func (s *strelets) UpdateReverseProxy(ctx context.Context, chains []*VirtualChain, ip string) error {
	return s.orchestrator.UpdateReverseProxy(ctx, getNginxConfig(chains, ip))
}
