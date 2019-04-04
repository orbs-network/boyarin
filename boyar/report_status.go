package boyar

import (
	"context"
	"fmt"
	"github.com/orbs-network/boyarin/log_types"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/scribe/log"
)

func ReportStatus(ctx context.Context, logger log.Logger) error {
	// We really don't need any options here since we're just observing
	orchestrator, err := adapter.NewDockerSwarm(adapter.OrchestratorOptions{})
	if err != nil {
		return err
	}
	defer orchestrator.Close()

	status, err := orchestrator.GetStatus(ctx)
	if err != nil {
		return fmt.Errorf("failed to report status: %s", err)
	}

	for _, s := range status {
		if s.Error != "" {
			logger.Error("service failure",
				log_types.VirtualChainId(getVcidFromServiceName(s.Name)),
				log.String("name", s.Name),
				log.String("state", s.State),
				log.Error(fmt.Errorf(s.Error)),
				log.String("logs", s.Logs))
		} else {
			logger.Info("service status",
				log_types.VirtualChainId(getVcidFromServiceName(s.Name)),
				log.String("name", s.Name),
				log.String("state", s.State),
				log.String("workerId", s.NodeID),
				log.String("createdAt", formatAsISO6801(s.CreatedAt)))
		}
	}

	if len(status) == 0 {
		logger.Info("no services found")
	}

	return nil
}