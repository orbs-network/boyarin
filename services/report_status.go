package services

import (
	"context"
	"encoding/json"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/utils"
	"github.com/orbs-network/boyarin/version"
	"github.com/orbs-network/govnr"
	"github.com/orbs-network/scribe/log"
	"io/ioutil"
	"time"
)

const SERVICE_STATUS_REPORT_PERIOD = 30 * time.Second
const SERVICE_STATUS_REPORT_TIMEOUT = 15 * time.Second

func WatchAndReportServicesStatus(ctx context.Context, logger log.Logger, statusFilePath string) govnr.ShutdownWaiter {
	errorHandler := utils.NewLogErrors("service status reporter", logger)
	return govnr.Forever(ctx, "service status reporter", errorHandler, func() {
		start := time.Now()
		ctxWithTimeout, cancel := context.WithTimeout(ctx, SERVICE_STATUS_REPORT_TIMEOUT)
		defer cancel()
		if status, err := GetStatus(ctxWithTimeout, logger, SERVICE_STATUS_REPORT_PERIOD); err != nil {
			logger.Error("status check failed", log.Error(err))
		} else {
			rawJSON, _ := json.MarshalIndent(status, "  ", "  ")
			if err := ioutil.WriteFile(statusFilePath, rawJSON, 0644); err != nil {
				logger.Error("failed to write status.json", log.Error(err))
			}
		}

		select {
		case <-ctx.Done():
		case <-time.After(SERVICE_STATUS_REPORT_PERIOD - time.Since(start)): // to report exactly every minute
		}
	})
}

type StatusResponse struct {
	Timestamp time.Time
	Status    string
	Error     string
	Payload   interface{}
}

func GetStatus(ctx context.Context, logger log.Logger, since time.Duration) (status StatusResponse, err error) {
	// We really don't need any options here since we're just observing
	orchestrator, err := adapter.NewDockerSwarm(adapter.OrchestratorOptions{}, logger)
	if err != nil {
		return
	}
	defer orchestrator.Close()

	containerStatus, err := orchestrator.GetStatus(ctx, since)
	if err != nil {
		status = StatusResponse{
			Status:    "Failed to query Docker Swarm",
			Timestamp: time.Now(),
			Error:     err.Error(),
			Payload: map[string]interface{}{
				"Version": version.GetVersion(),
			},
		}
	} else {
		services := make(map[string][]*adapter.ContainerStatus)
		for _, s := range containerStatus {
			services[s.Name] = append(services[s.Name], s)
		}

		status = StatusResponse{
			Status:    "OK",
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"Version":  version.GetVersion(),
				"Services": services,
			},
		}
	}

	return
}
