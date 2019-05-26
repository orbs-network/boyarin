package strelets

import (
	"context"
)

type UpdateServiceInput struct {
	Service *Service
}

func (s *strelets) UpdateService(ctx context.Context, input *UpdateServiceInput) error {
	panic("not implemented")
}
