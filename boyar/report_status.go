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
		fields := []*log.Field{
			log.String("name", s.Name),
			log.String("state", s.State),
			log.String("workerId", s.NodeID),
			log.String("createdAt", formatAsISO6801(s.CreatedAt)),
		}

		if vcid := getVcidFromServiceName(s.Name); vcid > 0 {
			fields = append(fields, log_types.VirtualChainId(vcid))
		}

		if s.Error != "" {
			fields = append(fields, log.Error(fmt.Errorf(s.Error)), log.String("logs", s.Logs))
			logger.Error("service failure", fields...)
		} else {
			logger.Info("service status", fields...)
		}
	}

	if len(status) == 0 {
		logger.Info("no services found")
	}

	return nil
}
