package strelets

import (
	"context"
	"fmt"
)

type ProvisionSharedNetworkInput struct {
	Name string
}

func (s *strelets) ProvisionSharedNetwork(ctx context.Context, input *ProvisionSharedNetworkInput) error {
	_, err := s.orchestrator.GetOverlayNetwork(ctx, input.Name)
	if err != nil {
		fmt.Errorf("failed to provision shared network %s: %s", input.Name, err)
	}
	return nil
}
