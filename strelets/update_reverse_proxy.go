package strelets

import "context"

func (s *strelets) UpdateReverseProxy(ctx context.Context, chains []*VirtualChain, ip string) error {
	if runner, err := s.orchestrator.PrepareReverseProxy(ctx, getNginxConfig(chains, ip)); err != nil {
		return err
	} else {
		return runner.Run(ctx)
	}
}
