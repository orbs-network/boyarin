package strelets

import (
	"context"
)

type ProvisionSharedNetworkInput struct {
	Name string
}

func (s *strelets) ProvisionSharedNetwork(ctx context.Context, input *ProvisionSharedNetworkInput) error {
	panic("removed")
}
