package strelets

import (
	"context"
	"github.com/orbs-network/boyarin/strelets/adapter"
)

type Strelets interface {
	UpdateService(ctx context.Context, input *UpdateServiceInput) error
	Orchestrator() adapter.Orchestrator
}

type strelets struct {
	orchestrator adapter.Orchestrator
}

func NewStrelets(docker adapter.Orchestrator) Strelets {
	return &strelets{
		orchestrator: docker,
	}
}

func (s *strelets) Orchestrator() adapter.Orchestrator {
	return s.orchestrator
}
