package strelets

import (
	"github.com/orbs-network/boyarin/strelets/adapter"
)

type Strelets interface {
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
