package services

import (
	"context"
	"fmt"
	"github.com/orbs-network/boyarin/log_types"
	"github.com/orbs-network/boyarin/strelets"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/utils"
	"github.com/orbs-network/govnr"
	"github.com/orbs-network/scribe/log"
	"time"
)

const SERVICE_STATUS_REPORT_PERIOD = 1 * time.Minute
const SERVICE_STATUS_REPORT_TIMEOUT = 30 * time.Second

func WatchAndReportServicesStatus(ctx context.Context, logger log.Logger) govnr.ShutdownWaiter {
	errorHandler := utils.NewLogErrors("service status reporter", logger)
	return govnr.Forever(ctx, "service status reporter", errorHandler, func() {
		start := time.Now()
		ctx2, cancel := context.WithTimeout(ctx, SERVICE_STATUS_REPORT_TIMEOUT)
		defer cancel()
		if err := reportStatus(ctx2, logger, SERVICE_STATUS_REPORT_PERIOD); err != nil {
			logger.Error("status check failed", log.Error(err))
		}

		select {
		case <-ctx.Done():
		case <-time.After(SERVICE_STATUS_REPORT_PERIOD - time.Since(start)): // to report exactly every minute
		}
	})
}

func formatAsISO6801(t time.Time) string {
	return t.Format(time.RFC3339)
}

func reportStatus(ctx context.Context, logger log.Logger, since time.Duration) error {
	// We really don't need any options here since we're just observing
	orchestrator, err := adapter.NewDockerSwarm(adapter.OrchestratorOptions{}, logger)
	if err != nil {
		return err
	}
	defer orchestrator.Close()

	status, err := orchestrator.GetStatus(ctx, since)
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

		if vcid := strelets.GetVcidFromServiceName(s.Name); vcid > 0 {
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
